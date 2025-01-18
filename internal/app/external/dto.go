package external

const (
	StatusRegistered Status = "REGISTERED"
	StatusInvalid    Status = "INVALID"
	StatusProcessing Status = "PROCESSING"
	StatusProcessed  Status = "PROCESSED"
)

type Status string

type AccrualResponse struct {
	Accrual *int   `json:"accrual,omitempty"`
	Order   string `json:"order"`
	Status  Status `json:"status"`
}
