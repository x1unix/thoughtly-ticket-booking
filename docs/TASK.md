# Thoughtly Application Development Engineer - Take Home Assignment

# General Guidelines

---

- **Due date:** 5 days from when you receive this assignment.
- **Readme:** Use GitHub with clear commit history. Include a top‑level `README.md` with run steps. Please provide a summary of trade‑offs and decisions made as well.
- **What we evaluate:** Code clarity, correctness, trade‑off thinking, availability + reliability (just describe how this would be achieved when the system scales), user experience, and documentation.
- **Testing:** comprehensive testing is a bonus but not required.

---

# Assignment - Ticket Booking System

### Goal

Build an end‑to‑end concert **ticket booking** app with a **React + TypeScript** frontend and a **Node.js** (TypeScript) backend. Prevent double‑booking, support a global user base (this means users can be globally distributed), and present a clean, functional UI (does not have to be fancy). Users can be mocked, we don’t need to see user management logic. This exercise is intentionally open-ended to see what technical design decisions you would make. 

### Functional Requirements

- **Ticket Catalog & Tiers:** VIP, Front Row, general admission (GA)
    - **Pricing:** VIP = $100, Front Row = $50, GA = $10
- **Availability:** UI to view all available tickets (required) and quantities per tier (optional)
- **Booking:** UI & API to book tickets (1+ quantity per tier)
- **No Double‑Booking:** Two users must not be able to book the **same ticket** at the same time.
- **Global Users:** Users may book from any country (assume a single currency display in USD).

### Non‑Functional Requirements

- **Availability target:** *four nines* (99.99%) **design intent** — you won’t implement HA multi‑region, just discuss how your design would achieve it in the Readme.
- **Scale assumptions:** ~1,000 DAU; peak ~500 concurrent users. Just discuss how your design would achieve it in the Readme.
- **Performance:** Booking request p95 < 500ms. Just discuss how your design would achieve it in the Readme.

### Constraints & Guidance

- **Language/Frameworks:** React + TS (frontend); Node.js + TS (backend) or Golang (backend)
- **Data store:** Use any transactional store. Postgres is fine. In‑memory stores are allowed **only** if you still demonstrate correct locking/idempotency.
- **Payments:** **Do not** integrate real payment providers— we can just simulate payment success/failure.
- **Auth:** No need for auth.

### Consistency & Concurrency (Critical)

Demonstrate how you prevent double‑booking **under race conditions. Discuss in the Readme and add comments in the code.** 

### Testing

- Up to you.

### Questions

- Reach out to alex@thoughtly.com
