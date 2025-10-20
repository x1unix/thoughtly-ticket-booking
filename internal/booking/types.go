package booking

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID   uuid.UUID `json:"id" db:"id"`
	Name string    `json:"name" db:"name"`
}

type TicketTier struct {
	TierID         uuid.UUID `json:"tier_id" db:"tier_id"`
	Name           string    `json:"name" db:"tier_name"`
	PriceCents     int       `json:"priceCents" db:"price_cents"`
	AvailableCount int       `json:"availableCount" db:"available_count"`
}

type CreateTierParams struct {
	PriceCents   int `json:"priceCents"`
	TicketsCount int `json:"ticketsCount"`
}

type EventCreateParams struct {
	EventName string `json:"name"`

	Tiers map[string]CreateTierParams `json:"tiers"`
}

type EventCreateResult struct {
	EventID uuid.UUID            `json:"eventId"`
	Tiers   map[string]uuid.UUID `json:"tiers"`
}

type ReservationParams struct {
	IdempotencyKey uuid.UUID          `json:"idempotencyKey"`
	ActorID        uuid.UUID          `json:"actorID"`
	EventID        uuid.UUID          `json:"eventID"`
	TicketsCount   map[uuid.UUID]uint `json:"ticketsCount"`
}

type ReservationResult struct {
	ReservationID uuid.UUID `json:"reservationID"`
	ExpiresAt     time.Time `json:"expiresAt"`
}
