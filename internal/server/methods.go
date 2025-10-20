package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/x1unix/thoughtly-ticket-booking/internal/booking"
)

func (srv *Server) handleCreateEvent(c *fiber.Ctx) error {
	var req booking.EventCreateParams
	if err := c.BodyParser(&req); err != nil {
		return errBadRequest("can't parse request: ", err)
	}

	rsp, err := srv.svc.CreateEvent(c.Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(rsp)
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

func (srv *Server) handleReserveTickets(c *fiber.Ctx) error {
	var params eventIDRequest
	if err := c.ParamsParser(&params); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	var body ReserveTicketsRequest
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	rsp, err := srv.svc.ReserveTickets(c.Context(), booking.ReservationParams{
		IdempotencyKey: body.IdempotencyKey,
		ActorID:        body.ActorID,
		EventID:        params.EventID,
		TicketsCount:   body.TicketsCount,
	})
	if err != nil {
		if booking.IsInsufficientTicketsError(err) {
			return errBadRequest(err)
		}

		return err
	}

	return c.JSON(rsp)
}

type userIDRequest struct {
	UserID uuid.UUID `params:"userID"`
}

func (srv *Server) handleListReservations(c *fiber.Ctx) error {
	var params userIDRequest
	if err := c.ParamsParser(&params); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	items, err := srv.svc.GetReservations(c.Context(), params.UserID)
	if err != nil {
		return err
	}

	return c.JSON(&ListReservationsResponse{
		Reservations: items,
	})
}

type reservationIDRequest struct {
	ReservationID uuid.UUID `params:"reservationID"`
}

func (srv *Server) handlePayReservation(c *fiber.Ctx) error {
	var params reservationIDRequest
	if err := c.ParamsParser(&params); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	var body booking.PaymentParams
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	rsp, err := srv.svc.PayReservation(c.Context(), body)
	if err != nil {
		if errors.Is(err, booking.ErrNotFound) {
			return errNotFound("reservation not found")
		}

		if errors.Is(err, booking.ErrReservationExpired) {
			return errBadRequest("reservation expired")
		}

		return err
	}

	return c.JSON(rsp)
}

func errNotFound(msg string) error {
	return fiber.NewError(http.StatusNotFound, msg)
}

func errBadRequest(args ...any) error {
	return fiber.NewError(http.StatusBadRequest, fmt.Sprint(args...))
}
