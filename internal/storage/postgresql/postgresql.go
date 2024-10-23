package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/madcarpet/gophermart/internal/logger"
	"github.com/madcarpet/gophermart/internal/models"
	"go.uber.org/zap"
)

func dbMigrate(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Log.Error("db driver error on migration", zap.Error(err))
		return err
	}
	m, err := migrate.NewWithDatabaseInstance("file://db/migrations", "postgres", driver)
	if err != nil {
		logger.Log.Error("migration instance creation error on migration", zap.Error(err))
		return err
	}
	_, dirty, err := m.Version()
	if err != nil {
		switch err {
		case migrate.ErrNilVersion:
			logger.Log.Info("no migration was applied yet - first migration")
		default:
			logger.Log.Error("checking database dirty on migration error", zap.Error(err))
			return err
		}

	}
	if dirty {
		logger.Log.Error("migration - database is in dirty state")
		return err
	}
	err = m.Up()
	if err != nil {
		switch err {
		case migrate.ErrNoChange:
			logger.Log.Info("migration - db version is up to date")
			return nil
		default:
			logger.Log.Error("db migration error", zap.Error(err))
			return err
		}

	}
	return nil
}

type PsqlStorage struct {
	dbAddresses string
	connection  *sql.DB
}

func NewPsqlStorage(dba string) *PsqlStorage {
	return &PsqlStorage{dbAddresses: dba}
}

func (s *PsqlStorage) InitStorage(ctx context.Context) error {
	db, err := sql.Open("pgx", s.dbAddresses)
	if err != nil {
		logger.Log.Error("openning db connection error", zap.Error(err))
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		logger.Log.Error("db ping err", zap.Error(err))
		return err
	}
	err = dbMigrate(db)
	if err != nil {
		return err
	}
	s.connection = db
	logger.Log.Info("db connection is ready")
	return nil
}

func (s *PsqlStorage) AddUser(ctx context.Context, user *models.User) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := s.connection.Begin()
	if err != nil {
		logger.Log.Error("add user error - transaction open failed", zap.Error(err))
		return err
	}
	_, err = tx.ExecContext(ctx,
		"INSERT INTO USERS (id,username,pwdhash) VALUES($1,$2,$3)",
		user.ID, user.Login, user.Password)
	if err != nil {
		tx.Rollback()
		logger.Log.Error("add user error - db inserting failed", zap.Error(err))
		return err
	}
	return tx.Commit()
}

func (s *PsqlStorage) IsLoginFree(ctx context.Context, login string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	loginReqResult, err := s.connection.QueryContext(ctx,
		"SELECT * FROM USERS WHERE username = $1", login)
	if err != nil {
		logger.Log.Error("is login free error - db getting loging rows failed", zap.Error(err))
		return true, err
	}
	defer loginReqResult.Close()

	if loginReqResult.Next() {
		return false, nil
	}
	if loginReqResult.Err() != nil {
		logger.Log.Error("is login free error - raws iterating error", zap.Error(err))
		return true, err
	}
	return true, nil
}

func (s *PsqlStorage) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	userReqResult := s.connection.QueryRowContext(ctx,
		"SELECT * FROM USERS WHERE id = $1", id)
	err := userReqResult.Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Log.Error("get user by id error - db row scan error", zap.Error(err))
		return nil, err
	}
	return &user, nil
}

func (s *PsqlStorage) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	var user models.User
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	userReqResult := s.connection.QueryRowContext(ctx,
		"SELECT * FROM USERS WHERE username = $1", login)
	err := userReqResult.Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Log.Error("get user by login error - db row scan error", zap.Error(err))
		return nil, err
	}
	return &user, nil
}

func (s *PsqlStorage) AddOrder(ctx context.Context, o models.Order) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := s.connection.Begin()
	if err != nil {
		logger.Log.Error("add order error - transaction open failed", zap.Error(err))
		return err
	}
	_, err = tx.ExecContext(ctx,
		"INSERT INTO orders (number,status_id,user_id, accrual,upload_at) VALUES($1,$2,$3,$4,$5)",
		o.Number, o.Status, o.UserID, o.Accrual, o.UploadedAt)
	if err != nil {
		tx.Rollback()
		logger.Log.Error("add order error - db inserting failed", zap.Error(err))
		return err
	}
	return tx.Commit()
}

func (s *PsqlStorage) AddOrderDelayed(ctx context.Context, num string, uid string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := s.connection.Begin()
	if err != nil {
		logger.Log.Error("add order delayed error - transaction open failed", zap.String("order", num), zap.Error(err))
		return err
	}
	_, err = tx.ExecContext(ctx,
		"INSERT INTO orders_delayed (number, user_id) VALUES($1,$2)",
		num, uid)
	if err != nil {
		tx.Rollback()
		logger.Log.Error("add order delayed error - db inserting failed", zap.String("order", num), zap.Error(err))
		return err
	}
	return tx.Commit()
}

func (s *PsqlStorage) GetOrderByNumber(ctx context.Context, num string) (*models.Order, error) {
	var order models.Order
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	orderReqResult := s.connection.QueryRowContext(ctx,
		"SELECT * FROM orders WHERE number = $1", num)
	err := orderReqResult.Scan(&order.Number, &order.Status, &order.UserID, &order.Accrual, &order.UploadedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Log.Error("get order by number error - db row scan error", zap.Error(err))
		return nil, err
	}
	return &order, nil
}

func (s *PsqlStorage) UpdateOrderStatus(ctx context.Context, num string, st string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := s.connection.Begin()
	if err != nil {
		logger.Log.Error("update order status - transaction open failed", zap.Error(err))
		return err
	}
	_, err = tx.ExecContext(ctx,
		"UPDATE orders SET status_id = $1 WHERE number = $2", st, num)
	if err != nil {
		tx.Rollback()
		logger.Log.Error("update order status error - db updating failed", zap.String("order", num), zap.Error(err))
		return err
	}
	return tx.Commit()
}

func (s *PsqlStorage) UpdateOrderAccrual(ctx context.Context, num string, ac float32) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := s.connection.Begin()
	if err != nil {
		logger.Log.Error("update order accrual - transaction open failed", zap.Error(err))
		return err
	}
	_, err = tx.ExecContext(ctx,
		"UPDATE orders SET accrual = $1 WHERE number = $2", ac, num)
	if err != nil {
		tx.Rollback()
		logger.Log.Error("update order accrual error - db updating failed", zap.String("order", num), zap.Error(err))
		return err
	}
	return tx.Commit()
}

func (s *PsqlStorage) GetOrdersDelayed(ctx context.Context, lim int) ([]models.OrderDelayed, error) {
	ordersDelayed := make([]models.OrderDelayed, 0, lim)
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	rows, err := s.connection.QueryContext(ctx,
		"SELECT * FROM orders_delayed LIMIT $1", lim)
	if err != nil {
		logger.Log.Error("get orders delayed error - query error", zap.Error(err))
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var o models.OrderDelayed
		err := rows.Scan(&o.Number, &o.UserID)
		if err != nil {
			logger.Log.Error("get orders delayed error - scan error", zap.Error(err))
			return nil, err
		}
		ordersDelayed = append(ordersDelayed, o)
	}
	if rows.Err() != nil {
		logger.Log.Error("get orders delayed error - rows iteration error", zap.Error(err))
		return nil, err
	}
	return ordersDelayed, nil
}

func (s *PsqlStorage) DeleteOrderDelayed(ctx context.Context, num string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, err := s.connection.ExecContext(ctx,
		"DELETE FROM orders_delayed WHERE number = $1", num)
	if err != nil {
		logger.Log.Error("delete order delayed error - delete error", zap.Error(err))
		return err
	}
	return nil
}

func (s *PsqlStorage) AddUserBalance(ctx context.Context, balance *models.UserBalance) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := s.connection.Begin()
	if err != nil {
		logger.Log.Error("add user balance error - transaction open failed", zap.Error(err))
		return err
	}
	_, err = tx.ExecContext(ctx,
		"INSERT INTO balance (id,user_id,current,withdrawn) VALUES($1,$2,$3,$4)",
		balance.ID, balance.UserID, balance.Current, balance.Withdrawn)
	if err != nil {
		tx.Rollback()
		logger.Log.Error("add user balance error - db inserting failed", zap.Error(err))
		return err
	}
	return tx.Commit()
}

func (s *PsqlStorage) UpdateCurrentBalance(ctx context.Context, c float32, uid string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := s.connection.Begin()
	if err != nil {
		logger.Log.Error("update current balance - transaction open failed", zap.Error(err))
		return err
	}
	_, err = tx.ExecContext(ctx,
		"UPDATE balance SET current = $1 WHERE user_id = $2", c, uid)
	if err != nil {
		tx.Rollback()
		logger.Log.Error("update current balance error - db updating failed", zap.Error(err))
		return err
	}
	return tx.Commit()
}

func (s *PsqlStorage) GetOrders(ctx context.Context, uid string) ([]models.Order, error) {
	orders := []models.Order{}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	rows, err := s.connection.QueryContext(ctx,
		`SELECT o.number, os.status_name, o.user_id, o.accrual, o.upload_at FROM orders o 
		LEFT JOIN order_status os ON o.status_id = os.status_id 
		WHERE o.user_id = $1`, uid)
	if err != nil {
		logger.Log.Error("get orders error - query error", zap.Error(err))
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var o models.Order
		err := rows.Scan(&o.Number, &o.Status, &o.UserID, &o.Accrual, &o.UploadedAt)
		if err != nil {
			logger.Log.Error("get orders error - scan error", zap.Error(err))
			return nil, err
		}
		orders = append(orders, o)
	}
	if rows.Err() != nil {
		logger.Log.Error("get orders error - iteration error", zap.Error(err))
		return nil, err
	}
	return orders, nil
}

func (s *PsqlStorage) GetBalanceByUserID(ctx context.Context, uid string) (*models.UserBalance, error) {
	var userBalance models.UserBalance
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	balanceReqResult := s.connection.QueryRowContext(ctx,
		"SELECT * FROM balance WHERE user_id = $1", uid)
	err := balanceReqResult.Scan(&userBalance.ID, &userBalance.UserID, &userBalance.Current, &userBalance.Withdrawn)
	if err != nil {
		logger.Log.Error("get balance by user id error - db row scan error", zap.Error(err))
		return nil, err
	}
	return &userBalance, nil
}

func (s *PsqlStorage) AddWithdrawal(ctx context.Context, withdraw *models.Withdrawals) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := s.connection.Begin()
	if err != nil {
		logger.Log.Error("add withdrawal error - transaction open failed", zap.Error(err))
		return err
	}
	_, err = tx.ExecContext(ctx,
		"INSERT INTO withdrawals (id,user_id,order_num,summ,processed_at) VALUES($1,$2,$3,$4,$5)",
		withdraw.ID, withdraw.UserID, withdraw.OrderNumber, withdraw.Sum, withdraw.ProcessedAt)
	if err != nil {
		tx.Rollback()
		logger.Log.Error("add withdrawal error - db inserting failed", zap.Error(err))
		return err
	}
	return tx.Commit()
}

func (s *PsqlStorage) GetWithdrawalByOrderNum(ctx context.Context, num string) (*models.Withdrawals, error) {
	var withdrawal models.Withdrawals
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	orderReqResult := s.connection.QueryRowContext(ctx,
		"SELECT * FROM withdrawals WHERE order_num = $1", num)
	err := orderReqResult.Scan(&withdrawal.ID, &withdrawal.UserID, &withdrawal.OrderNumber, &withdrawal.Sum, &withdrawal.ProcessedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Log.Error("get withdrawals by order number error - db row scan error", zap.Error(err))
		return nil, err
	}
	return &withdrawal, nil
}

func (s *PsqlStorage) GetWithdrawals(ctx context.Context, uid string) ([]models.Withdrawals, error) {
	withdrawals := []models.Withdrawals{}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	rows, err := s.connection.QueryContext(ctx,
		`SELECT * FROM withdrawals w 
		WHERE w.user_id = $1`, uid)
	if err != nil {
		logger.Log.Error("get withdrawals error - query error", zap.Error(err))
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var w models.Withdrawals
		err := rows.Scan(&w.ID, &w.UserID, &w.OrderNumber, &w.Sum, &w.ProcessedAt)
		if err != nil {
			logger.Log.Error("get withdrawals error - scan error", zap.Error(err))
			return nil, err
		}
		withdrawals = append(withdrawals, w)
	}
	if rows.Err() != nil {
		logger.Log.Error("get withdrawals error - iteration error", zap.Error(err))
		return nil, err
	}
	return withdrawals, nil
}

func (s *PsqlStorage) DBClose() error {
	return s.connection.Close()
}
