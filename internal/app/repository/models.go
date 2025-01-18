package repository

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Login  string
	Hash   string
	Salt   string
	UserID uuid.UUID
}

type Order struct {
	UploadedAt time.Time
	Accrual    *int
	Number     string
	Status     string
}

type Balance struct {
	Current   float64
	Withdrawn int
}

type Withdraw struct {
	ProcessedAt time.Time
	Order       string
	Sum         int
}
