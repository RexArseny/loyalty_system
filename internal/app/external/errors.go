package external

import "fmt"

type ErrTooManyRequests struct {
	retryAfter int
}

func NewErrTooManyRequests(retryAfter int) error {
	return &ErrTooManyRequests{
		retryAfter: retryAfter,
	}
}

func (e *ErrTooManyRequests) Error() string {
	return fmt.Sprintf("too many requests, retry after %d seconds", e.retryAfter)
}

func (e *ErrTooManyRequests) RetryAfter() int {
	return e.retryAfter
}
