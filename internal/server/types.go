package server

import "github.com/x1unix/thoughtly-ticket-booking/internal/booking"

type ErrorResponse struct {
	Error string `json:"error"`
}

type ListEventsResponse struct {
	Events []*booking.Event `json:"events"`
}

type ListTiersResponse struct {
	Tiers []*booking.TicketTier `json:"tiers"`
}
