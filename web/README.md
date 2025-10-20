# Ticket Booking Web App

A simple React TypeScript web application for booking event tickets.

## Features

- Browse available events
- View ticket tiers and pricing
- Reserve tickets for events
- View your reservations
- Process payments

## Setup

1. Install dependencies:
```bash
npm install
```

2. Start the development server:
```bash
npm run dev
```

The app will be available at http://localhost:3000

## API Configuration

The app is configured to proxy API requests to `http://localhost:8000`. Make sure the backend server is running before starting the web app.

## Tech Stack

- React 18
- TypeScript
- React Router v6
- Bootstrap 5
- Vite

## User ID

The app generates and stores a unique user ID in localStorage when you first use it. This ID is used to track your reservations.

