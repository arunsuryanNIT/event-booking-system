import { useState } from 'react'
import { Link } from 'react-router-dom'
import UserPicker from './UserPicker'

/**
 * Navbar renders the top navigation bar with links to
 * Events, My Bookings, Audit Log, and the user picker dropdown.
 * Includes a hamburger toggle for mobile viewports.
 */
export default function Navbar() {
  const [menuOpen, setMenuOpen] = useState(false)

  return (
    <nav className="navbar">
      <Link to="/" className="navbar-brand">Event Booking (Realtime)</Link>
      <button
        className="navbar-toggle"
        onClick={() => setMenuOpen((o) => !o)}
        aria-label="Toggle navigation"
      >
        <span className={`hamburger ${menuOpen ? 'open' : ''}`} />
      </button>
      <div className={`navbar-menu ${menuOpen ? 'navbar-menu--open' : ''}`}>
        <div className="navbar-links">
          <Link to="/" onClick={() => setMenuOpen(false)}>Events</Link>
          <Link to="/my-bookings" onClick={() => setMenuOpen(false)}>My Bookings</Link>
          <Link to="/audit" onClick={() => setMenuOpen(false)}>Audit Log</Link>
        </div>
        <UserPicker />
      </div>
    </nav>
  )
}
