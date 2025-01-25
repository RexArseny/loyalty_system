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
	Accrual    *float64
	Number     string
	Status     string
	UserID     uuid.UUID
}

type Balance struct {
	Current   float64
	Withdrawn float64
}

type Withdraw struct {
	ProcessedAt time.Time
	Order       string
	Sum         float64
}
