# Booking App

## Prerequisites 

* Go 1.25
* GNU Make
* Node JS (latest LTS) - to build UI
* Docker + Compose

## Bootstrapping

* Start containers using `docker-compose up -d`
* Apply database migrations using `make migrate-up`
* Start API server using `make run`
* Start React app:
  - `cd web && npm install && npm run dev`

## Notes

### Trade-Offs

To satisfy short time constraints but still guarantee double-booking prevention guarantees - application doesn't implement any kind of caching and fully relies on Postgres ACID guarantees.

### Hold TTL

Ticket hold status is stored as `hold_expires_at` timestamp column.

In ideal use case: instead of keeping ticket lock status in DB itself - we could use a redistributed lock using Redis.

This would:
- make ticket selection query more lightweight (atm app checks `hold_expires_at < now()` and query is not indexed).
- automatically remove expired holds by using `TTL` on keys.

### Booking stages

Each booking reservation stage - reserve and pay - are executed in isolated transations with *read commited* level.

During ticket reservation process, tickets that are already locked by another transaction (`FOR UPDATE SKIP LOCKED`) are skipped and server handles this case and returns an error to a client.


