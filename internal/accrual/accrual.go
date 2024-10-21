package accrual

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/madcarpet/gophermart/internal/constants"
	"github.com/madcarpet/gophermart/internal/logger"
	"github.com/madcarpet/gophermart/internal/models"
	"github.com/madcarpet/gophermart/internal/storage"
	"go.uber.org/zap"
)

type AccrualClient struct {
	url                   string
	storage               storage.Storage
	workers               int
	workersDelayed        int
	delay                 int
	delayOrdersQueryLimit int
	orderChan             <-chan *models.Order
	repeats               int
}

func NewAccrualClient(u string, s storage.Storage, w int, wd int, d int, doqlim int, ch <-chan *models.Order, r int) *AccrualClient {
	return &AccrualClient{url: u, storage: s, workers: w, workersDelayed: wd, delay: d, delayOrdersQueryLimit: doqlim, orderChan: ch, repeats: r}
}

func (c *AccrualClient) Start(ctx context.Context) {
	for i := range c.workers {
		go c.DealOrders(ctx, i)
	}
	for i := range c.workersDelayed {
		go c.DealOrdersDelayed(ctx, c.delay, c.delayOrdersQueryLimit, i)
	}
}

// Func to deal new orders
func (c *AccrualClient) DealOrders(ctx context.Context, wid int) {
	logger.Log.Info("accural client worker started", zap.Int("id", wid))
	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("accural client worker stopped by ctx", zap.Int("id", wid))
			return
		case order := <-c.orderChan:
			c.ProcessOrder(ctx, order.Number, order.UserId, false)
		}
	}
}

// Func to deal not proccessed, failed orders
func (c *AccrualClient) DealOrdersDelayed(ctx context.Context, delay int, lim int, wid int) {
	logger.Log.Info("accural client worker for delayed orders started", zap.Int("id", wid))
	tick := time.NewTicker(time.Duration(delay) * time.Second)
	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("accural client worker for delayed orders stopped by ctx", zap.Int("id", wid))
			return
		case <-tick.C:
			ordersDelayed, err := c.storage.GetOrdersDelayed(ctx, lim)
			if err != nil {
				logger.Log.Error("accural client worker get delayed orders error", zap.Error(err))
			}
			for _, order := range ordersDelayed {
				orderProcessed := c.ProcessOrder(ctx, order.Number, order.UserId, true)
				if orderProcessed {
					c.storage.DeleteOrderDelayed(ctx, order.Number)
				}
			}
		}
	}
}

func (c *AccrualClient) ProcessOrder(ctx context.Context, num string, uid string, delayed bool) (processed bool) {
	fullUrl := c.url + num
	processed = false
	saveToDB := true
	func(bool) {
		// retrying according to repeats
		for i := 1; i <= c.repeats; i++ {
			logger.Log.Debug("attepmt get accrual", zap.String("attempt", strconv.Itoa(i)), zap.String("order", num))
			//Send request
			r, err := http.Get(fullUrl)
			//If error repeat
			if err != nil {
				logger.Log.Error("get request to accrual failed", zap.String("order", num), zap.Error(err))
				continue
			}
			defer r.Body.Close()
			//Check statuscode
			switch r.StatusCode {
			//If no order in system save to DB
			case http.StatusNoContent:
				logger.Log.Debug("order not found in accrual system", zap.String("order", num))
				return
			//If too many requests wait for delay
			case http.StatusTooManyRequests:
				delay, err := strconv.Atoi(r.Header.Get("Retry-After"))
				if err != nil {
					logger.Log.Error("converting retry-after failed", zap.String("order", num), zap.Error(err))
					continue
				}
				delayTimer := time.NewTimer(time.Duration(delay) * time.Second)
				<-delayTimer.C
				//Drop repeat counter
				i = 1
				continue
			//Default reading body
			default:
				respBody, err := io.ReadAll(r.Body)
				if err != nil {
					//If error repeat
					logger.Log.Error("response body reading failed", zap.String("order", num), zap.Error(err))
					continue
				}
				//unmarshal body
				var orderInfo models.AccrualOrder
				err = json.Unmarshal(respBody, &orderInfo)
				if err != nil {
					//If error repeat
					logger.Log.Error("response body unmurshal failed", zap.String("order", num), zap.Error(err))
					continue
				}
				//Check order status
				switch orderInfo.OrderStatus {
				case "REGISTERED":
					//If registered and not processed exit and save to DB
					return
				case "PROCESSING":
					//If yet processing change status exit and save to DB
					err := c.storage.UpdateOrderStatus(ctx, num, constants.StatusProcessingOrder)
					if err != nil {
						//If error repeat
						logger.Log.Error("change status order error - got PROCESSING status from accrual system", zap.String("order", num), zap.Error(err))
						continue
					}
					return
				case "INVALID":
					//If invalid change status, exit and not save to DB
					err := c.storage.UpdateOrderStatus(ctx, num, constants.StatusInvalidOrder)
					if err != nil {
						//If error save to delayed dealing
						logger.Log.Error("change status order error - got INVALID status from accrual system", zap.String("order", num), zap.Error(err))
						return
					}
					saveToDB = false
					processed = true
					return
				case "PROCESSED":
					//If processed change status, update accrual and not save to DB
					err := c.storage.UpdateOrderStatus(ctx, num, constants.StatusProcessedOrder)
					if err != nil {
						//If error generate error log
						logger.Log.Error("change status order error - got PROCESSED status from accrual system", zap.String("order", num), zap.Error(err))
						return
					}
					err = c.storage.UpdateOrderAccrual(ctx, num, orderInfo.OrderAccrual)
					if err != nil {
						//If error generate error log
						logger.Log.Error("change accrual order error", zap.String("order", num), zap.Error(err))
						return
					}
					curBalance, err := c.storage.GetBalanceByUserId(ctx, uid)
					if err != nil {
						//If error generate error log
						logger.Log.Error("get balance error", zap.String("order", num), zap.Error(err))
						return
					}
					err = c.storage.UpdateCurrentBalance(ctx, curBalance.Current+orderInfo.OrderAccrual, uid)
					if err != nil {
						//If error generate error log
						logger.Log.Error("change balance error", zap.String("order", num), zap.Error(err))
						return
					}
					saveToDB = false
					processed = true
					return
				}
			}
		}
	}(saveToDB)
	switch delayed {
	case false:
		if saveToDB {
			err := c.storage.AddOrderDelayed(ctx, num, uid)
			if err != nil {
				logger.Log.Error("error adding order to delayed processing", zap.String("order", num), zap.Error(err))
			}
		}
	}
	return
}
