import { Link } from 'react-router-dom'
import UserPicker from './UserPicker'

/**
 * Navbar renders the top navigation bar with links to
 * Events, My Bookings, Audit Log, and the user picker dropdown.
 */
export default function Navbar() {
  return (
    <nav className="navbar">
      <Link to="/" className="navbar-brand">Event Booking</Link>
      <div className="navbar-links">
        <Link to="/">Events</Link>
        <Link to="/my-bookings">My Bookings</Link>
        <Link to="/audit">Audit Log</Link>
      </div>
      <UserPicker />
    </nav>
  )
}
