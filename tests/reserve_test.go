package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/x1unix/thoughtly-ticket-booking/internal/booking"
	"github.com/x1unix/thoughtly-ticket-booking/internal/server"
)

func TestTicketsReserve(t *testing.T) {
	eventName := fmt.Sprintf("CreateEventTest-%v", time.Now().UnixNano())
	tiers := map[string]booking.CreateTierParams{
		"VIP": {
			PriceCents:   100_00,
			TicketsCount: 50,
		},
		"Front Row": {
			PriceCents:   50_00,
			TicketsCount: 100,
		},
		"GA": {
			PriceCents:   10_00,
			TicketsCount: 1000,
		},
	}

	createRsp := client.CreateEvent(t, booking.EventCreateParams{
		EventName: eventName,
		Tiers:     tiers,
	})

	userID := uuid.New()
	eventID := createRsp.EventID
	tierIDs := createRsp.Tiers

	// Should fail if requested too many tickets
	_, err := client.ReserveTickets(eventID, server.ReserveTicketsRequest{
		IdempotencyKey: uuid.New(),
		ActorID:        userID,
		TicketsCount: map[uuid.UUID]uint{
			tierIDs["GA"]:  1,
			tierIDs["VIP"]: uint(tiers["VIP"].TicketsCount + 5),
		},
	})
	require.Error(t, err, "should fail if requested too much tickets")
	require.Contains(t, err.Error(), "not enough tickets available of tier")

	// Should success on enough tickets
	_, err = client.ReserveTickets(eventID, server.ReserveTicketsRequest{
		IdempotencyKey: uuid.New(),
		ActorID:        userID,
		TicketsCount: map[uuid.UUID]uint{
			tierIDs["VIP"]: uint(10),
		},
	})
	require.NoError(t, err)
}
