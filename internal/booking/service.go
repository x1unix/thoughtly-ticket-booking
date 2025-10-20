package booking

import (
	"context"
	"errors"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

var ErrNotFound = errors.New("not found")

type Service struct {
	db  *pgxpool.Pool
	rdb redis.UniversalClient
}

func NewService(db *pgxpool.Pool, rdb redis.UniversalClient) *Service {
	return &Service{
		db:  db,
		rdb: rdb,
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

func (svc Service) GetTicketTiers(ctx context.Context, eventID uuid.UUID) ([]*TicketTier, error) {
	var result []*TicketTier

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
	WHERE tt.event_id = $1  -- Replace with your event_id parameter
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
