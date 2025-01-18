package models

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type OrderResponse struct {
	Number     string `json:"number"`
	Status     string `json:"status"`
	Accrual    *int   `json:"accrual,omitempty"`
	UploadedAt string `json:"uploaded_at"`
}

type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn int     `json:"withdrawn"`
}

type WithdrawRequest struct {
	Order string `json:"order"`
	Sum   int    `json:"sum"`
}

type WithdrawResponse struct {
	Order       string `json:"order"`
	Sum         int    `json:"sum"`
	ProcessedAt string `json:"processed_at"`
}
