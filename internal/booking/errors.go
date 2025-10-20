package booking

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrNotFound           = errors.New("not found")
	ErrReservationExpired = errors.New("reservation is expired")
)

type InsufficientTicketsError struct {
	TierID uuid.UUID
}

func NewInsufficientTicketsError(tierID uuid.UUID) *InsufficientTicketsError {
	return &InsufficientTicketsError{
		TierID: tierID,
	}
}

func (err *InsufficientTicketsError) Error() string {
	return fmt.Sprintf("not enough tickets available of tier %q", err.TierID)
}

func IsInsufficientTicketsError(err error) bool {
	if err == nil {
		return false
	}

	e := &InsufficientTicketsError{}
	return errors.As(err, &e)
}
