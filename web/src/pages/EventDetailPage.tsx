import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { apiClient, TicketTier, ReserveTicketsRequest } from '../api/client';
import { ErrorAlert } from '../components/ErrorAlert';
import { SuccessAlert } from '../components/SuccessAlert';
import { LoadingSpinner } from '../components/LoadingSpinner';
import { formatPrice } from '../utils/format';
import { generateUUID } from '../utils/uuid';

export const EventDetailPage: React.FC = () => {
  const { eventId } = useParams<{ eventId: string }>();
  const navigate = useNavigate();
  const [tiers, setTiers] = useState<TicketTier[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [userId, setUserId] = useState<string>(() => {
    const stored = localStorage.getItem('userId');
    if (stored) return stored;
    const newId = generateUUID();
    localStorage.setItem('userId', newId);
    return newId;
  });
  const [ticketCounts, setTicketCounts] = useState<Record<string, number>>({});

  useEffect(() => {
    if (eventId) {
      loadTiers();
    }
  }, [eventId]);

  const loadTiers = async () => {
    if (!eventId) return;

    try {
      setLoading(true);
      setError(null);
      const data = await apiClient.getTicketTiers(eventId);
      setTiers(data);
      // Initialize ticket counts
      const counts: Record<string, number> = {};
      data.forEach((tier) => {
        counts[tier.tier_id] = 0;
      });
      setTicketCounts(counts);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load ticket tiers');
    } finally {
      setLoading(false);
    }
  };

  const handleCountChange = (tierId: string, value: number) => {
    setTicketCounts((prev) => ({
      ...prev,
      [tierId]: Math.max(0, value),
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!eventId) return;

    // Filter out tiers with 0 count
    const selectedTickets: Record<string, number> = {};
    Object.entries(ticketCounts).forEach(([tierId, count]) => {
      if (count > 0) {
        selectedTickets[tierId] = count;
      }
    });

    if (Object.keys(selectedTickets).length === 0) {
      setError('Please select at least one ticket');
      return;
    }

    const request: ReserveTicketsRequest = {
      idempotencyKey: generateUUID(),
      actorID: userId,
      ticketsCount: selectedTickets,
    };

    try {
      setSubmitting(true);
      setError(null);
      const result = await apiClient.reserveTickets(eventId, request);
      setSuccess(
        `Reservation created successfully! ID: ${result.reservationID}`
      );
      // Reset form
      const resetCounts: Record<string, number> = {};
      tiers.forEach((tier) => {
        resetCounts[tier.tier_id] = 0;
      });
      setTicketCounts(resetCounts);

      // Navigate to reservations after a delay
      setTimeout(() => {
        navigate('/reservations');
      }, 2000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to reserve tickets');
    } finally {
      setSubmitting(false);
    }
  };

  const getTotalPrice = (): number => {
    return tiers.reduce((total, tier) => {
      const count = ticketCounts[tier.tier_id] || 0;
      return total + tier.priceCents * count;
    }, 0);
  };

  const getTotalTickets = (): number => {
    return Object.values(ticketCounts).reduce((sum, count) => sum + count, 0);
  };

  if (loading) {
    return <LoadingSpinner />;
  }

  return (
    <div className="container">
      <div className="row">
        <div className="col-lg-8 mx-auto">
          <button
            className="btn btn-link ps-0 mb-3"
            onClick={() => navigate('/')}
          >
            ← Back to Events
          </button>

          <h1 className="mb-4">Book Tickets</h1>

          {error && (
            <ErrorAlert message={error} onClose={() => setError(null)} />
          )}

          {success && (
            <SuccessAlert message={success} onClose={() => setSuccess(null)} />
          )}

          <div className="card mb-4">
            <div className="card-body">
              <h6 className="text-muted mb-2">Your User ID</h6>
              <p className="font-monospace small">{userId}</p>
            </div>
          </div>

          {tiers.length === 0 ? (
            <div className="alert alert-info">
              No ticket tiers available for this event.
            </div>
          ) : (
            <form onSubmit={handleSubmit}>
              <div className="card mb-4">
                <div className="card-body">
                  <h5 className="card-title mb-4">Select Tickets</h5>

                  {tiers.map((tier) => (
                    <div key={tier.tier_id} className="mb-4">
                      <div className="d-flex justify-content-between align-items-start mb-2">
                        <div>
                          <h6 className="mb-1">{tier.name}</h6>
                          <p className="text-muted mb-1">
                            {formatPrice(tier.priceCents)} per ticket
                          </p>
                          <p className="text-muted small mb-0">
                            {tier.availableCount} available
                          </p>
                        </div>
                        <div className="input-group" style={{ width: '150px' }}>
                          <button
                            type="button"
                            className="btn btn-outline-secondary"
                            onClick={() =>
                              handleCountChange(
                                tier.tier_id,
                                (ticketCounts[tier.tier_id] || 0) - 1
                              )
                            }
                            disabled={!ticketCounts[tier.tier_id]}
                          >
                            -
                          </button>
                          <input
                            type="number"
                            className="form-control text-center"
                            value={ticketCounts[tier.tier_id] || 0}
                            onChange={(e) =>
                              handleCountChange(
                                tier.tier_id,
                                parseInt(e.target.value) || 0
                              )
                            }
                            min="0"
                            max={tier.availableCount}
                          />
                          <button
                            type="button"
                            className="btn btn-outline-secondary"
                            onClick={() =>
                              handleCountChange(
                                tier.tier_id,
                                (ticketCounts[tier.tier_id] || 0) + 1
                              )
                            }
                            disabled={
                              (ticketCounts[tier.tier_id] || 0) >=
                              tier.availableCount
                            }
                          >
                            +
                          </button>
                        </div>
                      </div>
                      <hr />
                    </div>
                  ))}
                </div>
              </div>

              {getTotalTickets() > 0 && (
                <div className="card mb-4">
                  <div className="card-body">
                    <h5 className="card-title">Order Summary</h5>
                    <div className="d-flex justify-content-between mb-2">
                      <span>Total Tickets:</span>
                      <strong>{getTotalTickets()}</strong>
                    </div>
                    <div className="d-flex justify-content-between">
                      <span>Total Price:</span>
                      <strong className="text-primary">
                        {formatPrice(getTotalPrice())}
                      </strong>
                    </div>
                  </div>
                </div>
              )}

              <button
                type="submit"
                className="btn btn-primary btn-lg w-100"
                disabled={submitting || getTotalTickets() === 0}
              >
                {submitting ? 'Reserving...' : 'Reserve Tickets'}
              </button>
            </form>
          )}
        </div>
      </div>
    </div>
  );
};

