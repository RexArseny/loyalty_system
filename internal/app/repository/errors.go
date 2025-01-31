package repository

import "fmt"

type ErrOriginalLoginUniqueViolation struct {
	login string
}

func NewErrOriginalLoginUniqueViolation(login string) error {
	return &ErrOriginalLoginUniqueViolation{
		login: login,
	}
}

func (e *ErrOriginalLoginUniqueViolation) Error() string {
	return fmt.Sprintf("login %s is not unique", e.login)
}

type ErrInvalidAuthData struct {
	login string
}

func NewErrInvalidAuthData(login string) error {
	return &ErrInvalidAuthData{
		login: login,
	}
}

func (e *ErrInvalidAuthData) Error() string {
	return fmt.Sprintf("login %s or password is incorrect", e.login)
}

type ErrAlreadyAdded struct {
	order string
}

func NewErrAlreadyAdded(order string) error {
	return &ErrAlreadyAdded{
		order: order,
	}
}

func (e *ErrAlreadyAdded) Error() string {
	return fmt.Sprintf("order %s has been already added", e.order)
}

type ErrAlreadyAddedByAnotherUser struct {
	order string
}

func NewErrAlreadyAddedByAnotherUser(order string) error {
	return &ErrAlreadyAddedByAnotherUser{
		order: order,
	}
}

func (e *ErrAlreadyAddedByAnotherUser) Error() string {
	return fmt.Sprintf("order %s has been already added by another user", e.order)
}

type ErrInvalidOrderNumber struct {
	order string
}

func NewErrInvalidOrderNumber(order string) error {
	return &ErrInvalidOrderNumber{
		order: order,
	}
}

func (e *ErrInvalidOrderNumber) Error() string {
	return fmt.Sprintf("invalid order number %s", e.order)
}
