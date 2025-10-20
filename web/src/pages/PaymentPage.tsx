import React, { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { apiClient, PaymentParams } from '../api/client';
import { ErrorAlert } from '../components/ErrorAlert';
import { SuccessAlert } from '../components/SuccessAlert';
import { formatPrice } from '../utils/format';

export const PaymentPage: React.FC = () => {
  const { reservationId } = useParams<{ reservationId: string }>();
  const navigate = useNavigate();
  const [cardNumber, setCardNumber] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!reservationId) return;

    if (!cardNumber || cardNumber.length < 13) {
      setError('Please enter a valid card number');
      return;
    }

    const params: PaymentParams = {
      reservationID: reservationId,
      cardNumber: cardNumber.replace(/\s/g, ''),
    };

    try {
      setSubmitting(true);
      setError(null);
      const result = await apiClient.payReservation(reservationId, params);
      setSuccess(
        `Payment successful! Transaction ID: ${result.txId}. Amount: ${formatPrice(
          result.amountCents
        )}`
      );
      setCardNumber('');

      // Navigate back to reservations after a delay
      setTimeout(() => {
        navigate('/reservations');
      }, 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Payment failed');
    } finally {
      setSubmitting(false);
    }
  };

  const formatCardNumber = (value: string): string => {
    const cleaned = value.replace(/\s/g, '');
    const groups = cleaned.match(/.{1,4}/g);
    return groups ? groups.join(' ') : cleaned;
  };

  const handleCardNumberChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value.replace(/\s/g, '');
    if (value.length <= 16 && /^\d*$/.test(value)) {
      setCardNumber(formatCardNumber(value));
    }
  };

  return (
    <div className="container">
      <div className="row">
        <div className="col-lg-6 mx-auto">
          <button
            className="btn btn-link ps-0 mb-3"
            onClick={() => navigate('/reservations')}
          >
            ← Back to Reservations
          </button>

          <h1 className="mb-4">Payment</h1>

          {error && (
            <ErrorAlert message={error} onClose={() => setError(null)} />
          )}

          {success && (
            <SuccessAlert message={success} onClose={() => setSuccess(null)} />
          )}

          <div className="card mb-4">
            <div className="card-body">
              <h6 className="text-muted mb-2">Reservation ID</h6>
              <p className="font-monospace small">{reservationId}</p>
            </div>
          </div>

          <div className="card">
            <div className="card-body">
              <h5 className="card-title mb-4">Payment Details</h5>

              <form onSubmit={handleSubmit}>
                <div className="mb-3">
                  <label htmlFor="cardNumber" className="form-label">
                    Card Number
                  </label>
                  <input
                    type="text"
                    className="form-control font-monospace"
                    id="cardNumber"
                    placeholder="1234 5678 9012 3456"
                    value={cardNumber}
                    onChange={handleCardNumberChange}
                    required
                    disabled={submitting}
                  />
                  <div className="form-text">
                    Test card: 4111 1111 1111 1111
                  </div>
                </div>

                <div className="alert alert-warning">
                  <small>
                    <strong>Note:</strong> This is a test payment system. Use
                    any test card number.
                  </small>
                </div>

                <button
                  type="submit"
                  className="btn btn-primary btn-lg w-100"
                  disabled={submitting}
                >
                  {submitting ? 'Processing...' : 'Pay Now'}
                </button>
              </form>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

