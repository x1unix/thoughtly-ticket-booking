package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/x1unix/thoughtly-ticket-booking/internal/booking"
)

func TestTicketsCreate(t *testing.T) {
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

	eventsRsp := client.GetEvents(t)
	require.Contains(t, eventsRsp.Events, &booking.Event{
		ID:   createRsp.EventID,
		Name: eventName,
	})

	tiersRsp := client.GetTicketTiers(t, createRsp.EventID)
	require.Len(t, tiersRsp.Tiers, len(tiers))
	for _, tier := range tiersRsp.Tiers {
		info, ok := tiers[tier.Name]
		require.Truef(t, ok, "missing tier %q", tier.Name)

		require.Equal(t, info, booking.CreateTierParams{
			PriceCents:   tier.PriceCents,
			TicketsCount: tier.AvailableCount,
		})
	}
}
