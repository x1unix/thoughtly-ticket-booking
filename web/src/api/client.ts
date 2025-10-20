// API Client for Ticket Booking System

export interface Event {
  id: string;
  name: string;
}

export interface TicketTier {
  tier_id: string;
  name: string;
  priceCents: number;
  availableCount: number;
}

export interface ReservationMeta {
  id: string;
  eventID: string;
  eventName: string;
  expiresAt: string;
  isPaid: boolean;
}

export interface ReserveTicketsRequest {
  idempotencyKey: string;
  actorID: string;
  ticketsCount: Record<string, number>;
}

export interface ReservationResult {
  reservationID: string;
  expiresAt: string;
}

export interface PaymentParams {
  reservationID: string;
  cardNumber: string;
}

export interface PaymentResult {
  txId: string;
  amountCents: number;
}

export interface ErrorResponse {
  error: string;
}

class ApiClient {
  private baseURL: string;

  constructor(baseURL: string = '/api') {
    this.baseURL = baseURL;
  }

  private async request<T>(
    path: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseURL}${path}`;
    const response = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });

    if (!response.ok) {
      const error: ErrorResponse = await response.json().catch(() => ({
        error: `HTTP ${response.status}: ${response.statusText}`,
      }));
      throw new Error(error.error);
    }

    return response.json();
  }

  async getEvents(): Promise<Event[]> {
    const data = await this.request<{ events: Event[] }>('/events');
    return data.events;
  }

  async getTicketTiers(eventID: string): Promise<TicketTier[]> {
    const data = await this.request<{ tiers: TicketTier[] }>(
      `/events/${eventID}/tiers`
    );
    return data.tiers;
  }

  async reserveTickets(
    eventID: string,
    params: ReserveTicketsRequest
  ): Promise<ReservationResult> {
    return this.request<ReservationResult>(`/events/${eventID}/reserve`, {
      method: 'POST',
      body: JSON.stringify(params),
    });
  }

  async getUserReservations(userID: string): Promise<ReservationMeta[]> {
    const data = await this.request<{ reservations: ReservationMeta[] }>(
      `/users/${userID}/reservations`
    );
    return data.reservations;
  }

  async payReservation(
    reservationID: string,
    params: PaymentParams
  ): Promise<PaymentResult> {
    return this.request<PaymentResult>(`/reservations/${reservationID}/payment`, {
      method: 'POST',
      body: JSON.stringify(params),
    });
  }
}

export const apiClient = new ApiClient();

