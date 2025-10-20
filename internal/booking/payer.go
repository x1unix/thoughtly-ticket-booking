package booking

import (
	"fmt"

	"github.com/google/uuid"
)

type PayResult struct {
	TXID uuid.UUID
}

type PayParams struct {
	ReservID    uuid.UUID
	Card        string
	AmountCents uint
}

type Payer interface {
	Pay(p PayParams) (*PayResult, error)
	Rollback(txID uuid.UUID) error
}

const KnownFakeCard = "1234567890"

type MockPayer struct{}

func (*MockPayer) Pay(p PayParams) (*PayResult, error) {
	if p.Card == KnownFakeCard {
		return &PayResult{
			TXID: uuid.New(),
		}, nil
	}

	return nil, fmt.Errorf("card is not in allowlist")
}

func (*MockPayer) Rollback(txID uuid.UUID) error {
	return nil
}
