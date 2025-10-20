package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/x1unix/thoughtly-ticket-booking/internal/booking"
)

func TestTicketsCreate(t *testing.T) {
	eventName := fmt.Sprintf("CreateEventTest-%v", time.Now().UnixNano())
	rsp := client.CreateEvent(t, booking.EventCreateParams{
		EventName: eventName,
		Tiers: map[string]booking.CreateTierParams{
			"VIP": {
				PriceCents:   100_00,
				TicketsCount: 10,
			},
			"Front Row": {
				PriceCents:   50_00,
				TicketsCount: 40,
			},
			"GA": {
				PriceCents:   10_00,
				TicketsCount: 100,
			},
		},
	})

	t.Logf("%#v", rsp)
}
