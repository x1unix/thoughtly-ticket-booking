import React from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { Navbar } from './components/Navbar';
import { EventsPage } from './pages/EventsPage';
import { EventDetailPage } from './pages/EventDetailPage';
import { ReservationsPage } from './pages/ReservationsPage';
import { PaymentPage } from './pages/PaymentPage';

export const App: React.FC = () => {
  return (
    <BrowserRouter>
      <Navbar />
      <Routes>
        <Route path="/" element={<EventsPage />} />
        <Route path="/events/:eventId" element={<EventDetailPage />} />
        <Route path="/reservations" element={<ReservationsPage />} />
        <Route path="/payment/:reservationId" element={<PaymentPage />} />
      </Routes>
    </BrowserRouter>
  );
};

