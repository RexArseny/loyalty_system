package models

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type OrderResponse struct {
	Number     string   `json:"number"`
	Status     string   `json:"status"`
	Accrual    *float64 `json:"accrual,omitempty"`
	UploadedAt string   `json:"uploaded_at"`
}

type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type WithdrawResponse struct {
	Order       string  `json:"order"`
	ProcessedAt string  `json:"processed_at"`
	Sum         float64 `json:"sum"`
}
