package server

import (
	"github.com/google/uuid"

	"github.com/x1unix/thoughtly-ticket-booking/internal/booking"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type ListEventsResponse struct {
	Events []*booking.Event `json:"events"`
}

type ListTiersResponse struct {
	Tiers []*booking.TicketTier `json:"tiers"`
}

type ReserveTicketsRequest struct {
	IdempotencyKey uuid.UUID          `json:"idempotencyKey"`
	ActorID        uuid.UUID          `json:"actorID"`
	TicketsCount   map[uuid.UUID]uint `json:"ticketsCount"`
}

type ListReservationsResponse struct {
	Reservations []*booking.ReservationMeta `json:"reservations"`
}
