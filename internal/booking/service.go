package booking

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const (
	reservationTTL = 15 * time.Minute
)

type Service struct {
	db    *pgxpool.Pool
	rdb   redis.UniversalClient
	payer Payer
}

func NewService(db *pgxpool.Pool, rdb redis.UniversalClient) *Service {
	return &Service{
		db:    db,
		rdb:   rdb,
		payer: &MockPayer{},
	}
}

// CreateEvent is test method used to create test events with tickets.
func (svc Service) CreateEvent(ctx context.Context, opts EventCreateParams) (result *EventCreateResult, err error) {
	eventID := uuid.New()
	tiers := make(map[string]uuid.UUID, len(opts.Tiers))

	tx, txErr := svc.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if txErr != nil {
		return nil, fmt.Errorf("can't open tx: %w", txErr)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	_, err = tx.Exec(ctx, `INSERT INTO events (id, name) VALUES ($1, $2)`, eventID, opts.EventName)
	if err != nil {
		return nil, fmt.Errorf("cannot insert event %q: %w", opts.EventName, err)
	}

	for k, v := range opts.Tiers {
		tierID := uuid.New()
		tiers[k] = tierID

		_, err := tx.Exec(
			ctx, `INSERT INTO ticket_tiers (id, event_id, name, price_cents) VALUES ($1, $2, $3, $4)`,
			tierID, eventID, k, v.PriceCents,
		)
		if err != nil {
			return nil, fmt.Errorf("can't create tier %q: %w", k, err)
		}

		_, err = tx.Exec(
			ctx,
			`INSERT INTO tickets (event_id, tier_id) SELECT $1::UUID, $2::UUID FROM generate_series(1, $3)`,
			eventID, tierID, v.TicketsCount,
		)
		if err != nil {
			return nil, fmt.Errorf("can't create tickets for tier %q: %w", k, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &EventCreateResult{
		EventID: eventID,
		Tiers:   tiers,
	}, nil
}

func (svc Service) GetEvents(ctx context.Context) ([]*Event, error) {
	var result []*Event
	err := pgxscan.Select(ctx, svc.db, &result, "SELECT id, name from events")
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return result, nil
}

func (svc Service) GetReservationEntries(ctx context.Context, reservationID uuid.UUID) (*ReservationMeta, error) {
	result := &ReservationMeta{}
	err := pgxscan.Get(
		ctx, svc.db, result,
		`SELECT r.id, r.expires_at, r.is_paid, e.event_id, e.name as event_name 
		FROM reservations r
		LEFT JOIN events e ON r.event_id = e.id
		WHERE r.id = $1
		LIMIT 1`,
		reservationID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("failed to query reservation: %w", err)
	}

	return result, nil
}

func (svc Service) GetReservations(ctx context.Context, userID uuid.UUID) ([]*ReservationMeta, error) {
	var result []*ReservationMeta
	err := pgxscan.Select(
		ctx, svc.db, &result,
		`SELECT r.id, r.expires_at, r.is_paid, e.id as event_id, e.name as event_name 
		FROM reservations r
		LEFT JOIN events e ON r.event_id = e.id
		WHERE r.actor_id = $1`,
		userID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to query reservations: %w", err)
	}

	return result, nil
}

type reservationHeader struct {
	ExpiresAt time.Time `db:"expires_at"`
	IsPaid    bool      `db:"is_paid"`
}

func (svc Service) PayReservation(ctx context.Context, params PaymentParams) (*PaymentResult, error) {
	rID := params.ReservationID
	cardNumber := params.CardNumber

	tx, err := svc.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open TX: %w", err)
	}

	defer tx.Rollback(ctx)

	h := &reservationHeader{}
	err = pgxscan.Get(ctx, tx, h, `SELECT expires_at, is_paid FROM reservations WHERE id = $1`, rID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("failed to get reservation: %w", err)
	}

	if h.IsPaid {
		return nil, errors.New("reservation already paid")
	}

	now := time.Now()
	if now.After(h.ExpiresAt) {
		return nil, ErrReservationExpired
	}

	// Get tickets by reservation ID and compute total price from ticket_tiers
	type ticketPrice struct {
		PriceCents int `db:"price_cents"`
	}

	var ticketPrices []ticketPrice
	err = pgxscan.Select(ctx, tx, &ticketPrices, `
		SELECT tt.price_cents
		FROM tickets t
		JOIN ticket_tiers tt ON t.tier_id = tt.id
		WHERE t.hold_token = $1
	`, rID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket prices: %w", err)
	}

	if len(ticketPrices) == 0 {
		return nil, errors.New("no tickets found for reservation")
	}

	// Compute total price
	var totalCents uint = 0
	for _, tp := range ticketPrices {
		totalCents += uint(tp.PriceCents)
	}

	// Call payer to process payment
	payResult, err := svc.payer.Pay(PayParams{
		ReservID:    rID,
		Card:        cardNumber,
		AmountCents: totalCents,
	})
	if err != nil {
		return nil, fmt.Errorf("payment failed: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE tickets 
		SET is_sold = true, hold_token = NULL, hold_expires_at = NULL
		WHERE hold_token = $1
	`, rID)
	if err != nil {
		_ = svc.payer.Rollback(payResult.TXID)
		return nil, fmt.Errorf("failed to mark tickets as sold: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE reservations 
		SET is_paid = true
		WHERE id = $1
	`, rID)
	if err != nil {
		_ = svc.payer.Rollback(payResult.TXID)
		return nil, fmt.Errorf("failed to mark reservation as paid: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		_ = svc.payer.Rollback(payResult.TXID)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &PaymentResult{
		TxID:        payResult.TXID,
		AmountCents: totalCents,
	}, nil
}

func (svc Service) ReserveTickets(ctx context.Context, params ReservationParams) (*ReservationResult, error) {
	reservationID := uuid.New()
	expireAt := time.Now().Add(reservationTTL)

	tx, err := svc.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open TX: %w", err)
	}

	// TODO: store and update reservation total price
	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx, `
		INSERT INTO reservations (id, event_id, actor_id, expires_at, idempotency_key)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (idempotency_key) DO NOTHING
		`,
		reservationID, params.EventID, params.ActorID, expireAt, params.IdempotencyKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create reservation: %w", err)
	}

	for tierID, qty := range params.TicketsCount {
		if qty == 0 {
			continue
		}

		rows, err := tx.Query(ctx, `
			WITH picked AS (
				SELECT id
				FROM tickets
				WHERE event_id = $1
					AND tier_id  = $2
					AND is_sold  = false
					AND (hold_expires_at IS NULL OR hold_expires_at < now())
				ORDER BY id
				FOR UPDATE SKIP LOCKED
				LIMIT $3
			)
			UPDATE tickets t
			SET hold_token = $4, hold_expires_at = $5
			WHERE t.id IN (SELECT id FROM picked)
			RETURNING t.id
		`,
			params.EventID, tierID, qty, reservationID, expireAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to lock tickets of tier %q: %w", tierID, err)
		}

		// cmp locked and expected count
		got := 0
		for rows.Next() {
			got++
		}
		rows.Close()

		if got != int(qty) {
			return nil, NewInsufficientTicketsError(tierID)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &ReservationResult{
		ReservationID: reservationID,
		ExpiresAt:     expireAt,
	}, nil
}

func (svc Service) GetTicketTiers(ctx context.Context, eventID uuid.UUID) ([]*TicketTier, error) {
	var result []*TicketTier

	// TODO: use different approach to check counters as query is quite expensive.
	query := `
	SELECT 
		tt.id AS tier_id,
		tt.name AS tier_name,
		tt.price_cents,
		COUNT(*) FILTER (
			WHERE t.is_sold = FALSE 
			AND (t.hold_expires_at IS NULL OR t.hold_expires_at < now())
		) AS available_count
	FROM ticket_tiers tt
	LEFT JOIN tickets t ON t.tier_id = tt.id
	WHERE tt.event_id = $1
	GROUP BY tt.id, tt.name, tt.price_cents, tt.event_id
	ORDER BY tt.price_cents
`
	err := pgxscan.Select(ctx, svc.db, &result, query, eventID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return result, err
}
