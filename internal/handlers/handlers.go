package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/madcarpet/gophermart/internal/authorization"
	"github.com/madcarpet/gophermart/internal/constants"
	"github.com/madcarpet/gophermart/internal/logger"
	"github.com/madcarpet/gophermart/internal/middlewares"
	"github.com/madcarpet/gophermart/internal/models"
	"github.com/madcarpet/gophermart/internal/storage"
	"github.com/madcarpet/gophermart/internal/utils"
	"go.uber.org/zap"
)

func RegisterPostHandler(ctx context.Context, s storage.Storage, a authorization.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//Content-Type check
		if appType := r.Header.Get("Content-Type"); appType != constants.CntTypeHeaderJson {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Wrong request Content-Type"))
			return
		}
		//Getting body data
		reqBody, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			logger.Log.Error("registration handler error - body reading error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		var userData models.User
		var userBal models.UserBalance
		//Deserialisation
		if err = json.Unmarshal(reqBody, &userData); err != nil {
			logger.Log.Debug("registration handler error - body deserialisation error", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Wrong request format"))
			return
		}
		//UUID generation
		userData.Id = uuid.New().String()
		userBal.Id = uuid.New().String()
		userBal.UserId = userData.Id
		//PWD hashing
		pwdHash := sha256.Sum256([]byte(userData.Password))
		pwdHashString := hex.EncodeToString(pwdHash[:])
		userData.Password = pwdHashString
		// Checking user existance
		checkLoginFree, err := s.IsLoginFree(ctx, userData.Login)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		if !checkLoginFree {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("Login already used"))
			return
		}
		// Adding user
		if err := s.AddUser(ctx, &userData); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		// Adding balance
		if err := s.AddUserBalance(ctx, &userBal); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		// Producing token
		token, err := a.ProduceToken(userData.Id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		w.Header().Add(constants.CookieToken, token)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("User registratered and authorized successfully"))
	}
}

func LoginPostHandler(ctx context.Context, s storage.Storage, a authorization.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//Content-Type check
		if appType := r.Header.Get("Content-Type"); appType != constants.CntTypeHeaderJson {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Wrong request Content-Type"))
			return
		}
		//Getting body data
		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Log.Error("authorization handler error - body reading error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		var userData models.User
		//Deserialisation
		if err = json.Unmarshal(reqBody, &userData); err != nil {
			logger.Log.Debug("authorization handler error - body deserialisation error", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Wrong request format"))
			return
		}
		//PWD hashing
		pwdHash := sha256.Sum256([]byte(userData.Password))
		pwdHashString := hex.EncodeToString(pwdHash[:])
		userData.Password = pwdHashString
		//Get user from DB
		user, err := s.GetUserByLogin(ctx, userData.Login)
		if err != nil {
			logger.Log.Error("authorization handler error - getting user by login error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		// Check user existance
		if user == nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Wrong username or password"))
			return
		}
		// Check password hash validity
		if pwdHashString != user.Password {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Wrong username or password"))
			return
		}
		// Prodicing token
		token, err := a.ProduceToken(user.Id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		w.Header().Add(constants.CookieToken, token)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("User authorized successfully"))
	}
}

func OrdersGetHandler(ctx context.Context, s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//UID after authorization getting
		userID := r.Context().Value(middlewares.Uid)
		userIDStr, ok := userID.(string)
		if !ok {
			logger.Log.Error("orders get handler error - getting uuid value from request context failed")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		// get orders from storage
		orders, err := s.GetOrders(ctx, userIDStr)
		if err != nil {
			logger.Log.Error("orders get handler error - getting orders grom database failed")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		// if len of orders is equal to 0 return 204
		if len(orders) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// deserialization
		jsonOrders, err := json.Marshal(orders)
		if err != nil {
			logger.Log.Error("orders get handler error - orders deserialization failed")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		w.Header().Set("Content-Type", constants.CntTypeHeaderJson)
		w.WriteHeader(http.StatusOK)
		w.Write(jsonOrders)
	}
}

func OrdersPostHandler(ctx context.Context, s storage.Storage, ch chan<- *models.Order) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//Content-Type check
		if appType := r.Header.Get("Content-Type"); appType != constants.CntTypeHeaderText {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Wrong request Content-Type"))
			return
		}
		//UID after authorization getting
		userID := r.Context().Value(middlewares.Uid)
		userIDStr, ok := userID.(string)
		if !ok {
			logger.Log.Error("orders post handler error - getting uuid value from request context failed")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		//Getting body data
		reqBody, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			logger.Log.Error("add order handler error - body reading error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		//Check if request body empty
		orderNum := string(reqBody)
		if orderNum == "" {
			logger.Log.Debug("add order handler error  - body is empty")
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("Request body is empty"))
			return
		}
		// Trim whitespace from request
		orderNum = strings.TrimSpace(orderNum)
		orderNum = strings.ReplaceAll(orderNum, " ", "")
		//Check if order string is numbers only
		if ok = utils.CheckIsNumbersOnly(orderNum); !ok {
			logger.Log.Debug("add order handler error  - bad order format")
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("Order format is incorrect"))
			return
		}
		//Check Luhn
		if ok = utils.CheckLuhn(orderNum); !ok {
			logger.Log.Debug("add order handler error  - luhn algo check failed")
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("Order format is incorrect"))
			return
		}
		// Check oreder existance
		orderCheckResult, err := s.GetOrderByNumber(ctx, orderNum)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		if orderCheckResult != nil {
			switch orderCheckResult.UserId {
			case userIDStr:
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Order already uploaded"))
				return
			default:
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte("Order already exist in system"))
				return
			}
		}
		orderTime := time.Now()
		orderTimeStr := orderTime.Format(time.RFC3339)
		order := models.Order{Number: orderNum, Status: constants.StatusNewOrder, UserId: userIDStr, UploadedAt: orderTimeStr}
		if err = s.AddOrder(ctx, order); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		// Put order in oreders channel for workers
		ch <- &order
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Order accepted"))

	}
}

func BalanceGetHandler(ctx context.Context, s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//UID after authorization getting
		userID := r.Context().Value(middlewares.Uid)
		userIDStr, ok := userID.(string)
		if !ok {
			logger.Log.Error("balance get handler error - getting uuid value from request context failed")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		// get balance from storage
		balance, err := s.GetBalanceByUserId(ctx, userIDStr)
		if err != nil {
			logger.Log.Error("balance get handler error - getting orders grom database failed")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		// deserialization
		jsonBalance, err := json.Marshal(balance)
		if err != nil {
			logger.Log.Error("balance get handler error - orders deserialization failed")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		w.Header().Set("Content-Type", constants.CntTypeHeaderJson)
		w.WriteHeader(http.StatusOK)
		w.Write(jsonBalance)
	}
}

func WithdrawPostHandler(ctx context.Context, s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//Content-Type check
		if appType := r.Header.Get("Content-Type"); appType != constants.CntTypeHeaderJson {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Wrong request Content-Type"))
			return
		}
		//UID after authorization getting
		userID := r.Context().Value(middlewares.Uid)
		userIDStr, ok := userID.(string)
		if !ok {
			logger.Log.Error("withdraw handler error - getting uuid value from request context failed")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		//Getting body data
		reqBody, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			logger.Log.Error("withdraw handler error - body reading error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		var withdraw models.Withdrawals
		//Deserialisation
		if err = json.Unmarshal(reqBody, &withdraw); err != nil {
			logger.Log.Debug("withdraw handler error - body deserialisation error", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Wrong request format"))
			return
		}
		//Check if order is empty
		if withdraw.OrderNumber == "" {
			logger.Log.Debug("withdraw handler error  - order is empty")
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("Order is empty"))
			return
		}
		// Trim whitespace from order
		withdraw.OrderNumber = strings.TrimSpace(withdraw.OrderNumber)
		withdraw.OrderNumber = strings.ReplaceAll(withdraw.OrderNumber, " ", "")

		//Check if order string is numbers only
		if ok := utils.CheckIsNumbersOnly(withdraw.OrderNumber); !ok {
			logger.Log.Debug("withdraw handler error  - bad order format")
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("Order format is incorrect"))
			return
		}
		//Check Luhn
		if ok := utils.CheckLuhn(withdraw.OrderNumber); !ok {
			logger.Log.Debug("withdraw handler error  - luhn algo check failed")
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("Order format is incorrect"))
			return
		}
		//Check if order already processed
		orderExists, err := s.GetWithdrawalByOrderNum(ctx, withdraw.OrderNumber)
		if err != nil {
			logger.Log.Error("withdraw handler error - check order existance error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		if orderExists != nil {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("Order already processed"))
			return
		}
		// Check withdraw sum > 0
		if withdraw.Sum <= 0 {
			logger.Log.Debug("withdraw handler error  - sum is negative or zero")
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("Sum is incorrect"))
			return
		}
		// Get user balance
		balance, err := s.GetBalanceByUserId(ctx, userIDStr)
		if err != nil {
			logger.Log.Error("withdraw handler error - get user balance error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		//Check if balance is enough
		if balance.Current < withdraw.Sum {
			w.WriteHeader(http.StatusPaymentRequired)
			w.Write([]byte("Not enough credits on balance"))
			return
		}
		//Update balance
		newBalance := balance.Current - withdraw.Sum
		err = s.UpdateCurrentBalance(ctx, newBalance, userIDStr)
		if err != nil {
			logger.Log.Error("withdraw handler error - update balance error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}

		//Register withdraw
		withdraw.Id = uuid.New().String()
		withdraw.UserId = userIDStr
		orderTime := time.Now()
		withdraw.ProcessedAt = orderTime.Format(time.RFC3339)
		err = s.AddWithdrawal(ctx, &withdraw)
		if err != nil {
			logger.Log.Error("withdraw handler error - add withdrawal to db error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("withdrawal accepted"))
	}
}

func WithdrawlsGetHandler(ctx context.Context, s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//UID after authorization getting
		userID := r.Context().Value(middlewares.Uid)
		userIDStr, ok := userID.(string)
		if !ok {
			logger.Log.Error("withdrawals get handler error - getting uuid value from request context failed")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		// get withdrawals from storage
		withdrawals, err := s.GetWithdrawals(ctx, userIDStr)
		if err != nil {
			logger.Log.Error("orders get handler error - getting withdrawals from database failed")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		// if len of withdrawals is equal to 0 return 204
		if len(withdrawals) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// deserialization
		jsonWithdrawals, err := json.Marshal(withdrawals)
		if err != nil {
			logger.Log.Error("withdrawals get handler error - withdrawals deserialization failed")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		w.Header().Set("Content-Type", constants.CntTypeHeaderJson)
		w.WriteHeader(http.StatusOK)
		w.Write(jsonWithdrawals)
	}
}

func NotFoundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Gophermart page not found"))
	}
}
