import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { apiClient, Event } from '../api/client';
import { ErrorAlert } from '../components/ErrorAlert';
import { LoadingSpinner } from '../components/LoadingSpinner';

export const EventsPage: React.FC = () => {
  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadEvents();
  }, []);

  const loadEvents = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await apiClient.getEvents();
      setEvents(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load events');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <LoadingSpinner />;
  }

  return (
    <div className="container">
      <h1 className="mb-4">Available Events</h1>

      {error && <ErrorAlert message={error} onClose={() => setError(null)} />}

      {events.length === 0 ? (
        <div className="alert alert-info">
          No events available at the moment.
        </div>
      ) : (
        <div className="row">
          {events.map((event) => (
            <div key={event.id} className="col-md-6 col-lg-4 mb-4">
              <div className="card h-100">
                <div className="card-body">
                  <h5 className="card-title">{event.name}</h5>
                  <p className="card-text text-muted">Event ID: {event.id}</p>
                  <Link
                    to={`/events/${event.id}`}
                    className="btn btn-primary w-100"
                  >
                    View Details & Book
                  </Link>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

