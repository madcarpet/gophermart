package models

type User struct {
	ID       string
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Order struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	UserID     string  `json:"-"`
	Accrual    float32 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

type OrderDelayed struct {
	Number string
	UserID string
}

type AccrualOrder struct {
	OrderNumber  string  `json:"order"`
	OrderStatus  string  `json:"status"`
	OrderAccrual float32 `json:"accrual"`
}

type UserBalance struct {
	ID        string  `json:"-"`
	UserID    string  `json:"-"`
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

type Withdrawals struct {
	ID          string  `json:"-"`
	UserID      string  `json:"-"`
	OrderNumber string  `json:"order"`
	Sum         float32 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}
