package storage

import (
	"context"

	"github.com/madcarpet/gophermart/internal/models"
)

type Storage interface {
	InitStorage(ctx context.Context) error
	DbClose() error

	AddUser(ctx context.Context, user *models.User) error
	IsLoginFree(ctx context.Context, login string) (bool, error)
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
	AddOrder(ctx context.Context, o models.Order) error
	GetOrderByNumber(ctx context.Context, id string) (*models.Order, error)
	AddUserBalance(ctx context.Context, balance *models.UserBalance) error
	UpdateCurrentBalance(ctx context.Context, c float32, uid string) error
	GetOrders(ctx context.Context, uid string) ([]models.Order, error)
	GetBalanceByUserId(ctx context.Context, uid string) (*models.UserBalance, error)
	AddWithdrawal(ctx context.Context, withdraw *models.Withdrawals) error
	GetWithdrawalByOrderNum(ctx context.Context, num string) (*models.Withdrawals, error)
	GetWithdrawals(ctx context.Context, uid string) ([]models.Withdrawals, error)

	AddOrderDelayed(ctx context.Context, num string, uid string) error
	GetOrdersDelayed(ctx context.Context, lim int) ([]models.OrderDelayed, error)
	UpdateOrderStatus(ctx context.Context, num string, st string) error
	UpdateOrderAccrual(ctx context.Context, num string, ac float32) error
	DeleteOrderDelayed(ctx context.Context, num string) error
}
