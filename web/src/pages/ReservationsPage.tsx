import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { apiClient, ReservationMeta } from '../api/client';
import { ErrorAlert } from '../components/ErrorAlert';
import { LoadingSpinner } from '../components/LoadingSpinner';
import { formatDateTime } from '../utils/format';

export const ReservationsPage: React.FC = () => {
  const [reservations, setReservations] = useState<ReservationMeta[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [userId] = useState<string>(() => {
    return localStorage.getItem('userId') || '';
  });

  useEffect(() => {
    if (userId) {
      loadReservations();
    } else {
      setLoading(false);
    }
  }, [userId]);

  const loadReservations = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await apiClient.getUserReservations(userId);
      setReservations(data);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'Failed to load reservations'
      );
    } finally {
      setLoading(false);
    }
  };

  const isExpired = (expiresAt: string): boolean => {
    return new Date(expiresAt) < new Date();
  };

  if (loading) {
    return <LoadingSpinner />;
  }

  if (!userId) {
    return (
      <div className="container">
        <div className="alert alert-warning">
          No user ID found. Please make a reservation first.
        </div>
        <Link to="/" className="btn btn-primary">
          Browse Events
        </Link>
      </div>
    );
  }

  return (
    <div className="container">
      <h1 className="mb-4">My Reservations</h1>

      <div className="card mb-4">
        <div className="card-body">
          <h6 className="text-muted mb-2">Your User ID</h6>
          <p className="font-monospace small mb-0">{userId}</p>
        </div>
      </div>

      {error && <ErrorAlert message={error} onClose={() => setError(null)} />}

      {reservations.length === 0 ? (
        <div className="alert alert-info">
          <p className="mb-3">You don't have any reservations yet.</p>
          <Link to="/" className="btn btn-primary">
            Browse Events
          </Link>
        </div>
      ) : (
        <div className="row">
          {reservations.map((reservation) => {
            const expired = isExpired(reservation.expiresAt);
            return (
              <div key={reservation.id} className="col-lg-6 mb-4">
                <div className="card h-100">
                  <div className="card-body">
                    <div className="d-flex justify-content-between align-items-start mb-3">
                      <h5 className="card-title mb-0">
                        {reservation.eventName}
                      </h5>
                      {reservation.isPaid ? (
                        <span className="badge bg-success">Paid</span>
                      ) : expired ? (
                        <span className="badge bg-danger">Expired</span>
                      ) : (
                        <span className="badge bg-warning text-dark">
                          Pending
                        </span>
                      )}
                    </div>

                    <p className="text-muted small mb-2">
                      <strong>Reservation ID:</strong>
                      <br />
                      <span className="font-monospace">{reservation.id}</span>
                    </p>

                    <p className="text-muted small mb-2">
                      <strong>Expires:</strong>{' '}
                      {formatDateTime(reservation.expiresAt)}
                    </p>

                    {!reservation.isPaid && !expired && (
                      <Link
                        to={`/payment/${reservation.id}`}
                        className="btn btn-primary w-100 mt-3"
                      >
                        Pay Now
                      </Link>
                    )}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};

