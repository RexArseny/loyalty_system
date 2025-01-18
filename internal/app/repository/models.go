package repository

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserID uuid.UUID
	Login  string
	Hash   string
	Salt   string
}

type Order struct {
	Number     string
	Status     string
	Accrual    *int
	UploadedAt time.Time
}

type Balance struct {
	Current   float64
	Withdrawn int
}

type Withdraw struct {
	Order       string
	Sum         int
	ProcessedAt time.Time
}
