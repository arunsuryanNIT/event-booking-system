import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { UserProvider } from './context/UserContext'
import Navbar from './components/Navbar'
import HomePage from './pages/HomePage'
import EventPage from './pages/EventPage'
import MyBookingsPage from './pages/MyBookingsPage'
import AuditLogPage from './pages/AuditLogPage'

export default function App() {
  return (
    <UserProvider>
      <BrowserRouter>
        <Navbar />
        <div className="container">
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/events/:id" element={<EventPage />} />
            <Route path="/my-bookings" element={<MyBookingsPage />} />
            <Route path="/audit" element={<AuditLogPage />} />
          </Routes>
        </div>
      </BrowserRouter>
    </UserProvider>
  )
}
