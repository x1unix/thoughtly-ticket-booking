package server

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/x1unix/thoughtly-ticket-booking/internal/booking"
)

func (srv *Server) handleCreateEvent(c *fiber.Ctx) error {
	var req booking.EventCreateParams
	if err := c.BodyParser(&req); err != nil {
		return errBadRequest(err, "can't parse request")
	}

	rsp, err := srv.svc.CreateEvent(c.Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(rsp)
}

type ListEventsResponse struct {
	Events []*booking.Event `json:"events"`
}

func (srv *Server) handleListEvents(c *fiber.Ctx) error {
	items, err := srv.svc.GetEvents(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(ListEventsResponse{
		Events: items,
	})
}

type eventIDRequest struct {
	EventID uuid.UUID `params:"eventID"`
}

type ListTiersResponse struct {
	Tiers []*booking.TicketTier `json:"tiers"`
}

func (srv *Server) handleListTiersSummary(c *fiber.Ctx) error {
	var params eventIDRequest
	if err := c.ParamsParser(&params); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	items, err := srv.svc.GetTicketTiers(c.Context(), params.EventID)
	if err != nil {
		if errors.Is(err, booking.ErrNotFound) {
			return errNotFound("event not found")
		}

		return err
	}

	return c.JSON(ListTiersResponse{
		Tiers: items,
	})
}

func errNotFound(msg string) error {
	return fiber.NewError(http.StatusNotFound, msg)
}

func errBadRequest(err error, msg string) error {
	return fiber.NewError(http.StatusBadRequest, msg, err.Error())
}
