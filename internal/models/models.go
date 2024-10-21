package models

type User struct {
	Id       string
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Order struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	UserId     string  `json:"-"`
	Accrual    float32 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

type OrderDelayed struct {
	Number string
	UserId string
}

type AccrualOrder struct {
	OrderNumber  string  `json:"order"`
	OrderStatus  string  `json:"status"`
	OrderAccrual float32 `json:"accrual"`
}

type UserBalance struct {
	Id        string  `json:"-"`
	UserId    string  `json:"-"`
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

type Withdrawals struct {
	Id          string  `json:"-"`
	UserId      string  `json:"-"`
	OrderNumber string  `json:"order"`
	Sum         float32 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}
