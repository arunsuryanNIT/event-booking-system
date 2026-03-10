import { useState } from 'react'
import { cancelBooking } from '../api/client'
import { useUser } from '../context/UserContext'

/**
 * MyBookings renders a list of the current user's bookings
 * with a Cancel button on each active booking.
 *
 * @param {{ bookings: Array, onCancelled: Function }} props
 */
export default function MyBookings({ bookings, onCancelled }) {
  const { currentUser } = useUser()
  const [cancellingId, setCancellingId] = useState(null)
  const [message, setMessage] = useState(null)

  const handleCancel = async (bookingId) => {
    if (!currentUser) return
    setCancellingId(bookingId)
    setMessage(null)
    try {
      await cancelBooking(bookingId, currentUser.id)
      setMessage({ type: 'success', text: 'Booking cancelled.' })
      if (onCancelled) onCancelled()
    } catch (err) {
      setMessage({ type: 'error', text: err.error || 'Cancel failed' })
    } finally {
      setCancellingId(null)
    }
  }

  if (!bookings.length) {
    return <p className="empty-state">No bookings yet.</p>
  }

  return (
    <div>
      {message && (
        <p className={`msg msg-${message.type}`}>{message.text}</p>
      )}
      <div className="table-wrap">
      <table className="table">
        <thead>
          <tr>
            <th>Event</th>
            <th>Status</th>
            <th>Booked At</th>
            <th>Action</th>
          </tr>
        </thead>
        <tbody>
          {bookings.map((b) => (
            <tr key={b.id}>
              <td>{b.event_title}</td>
              <td>
                <span className={`badge badge-${b.status}`}>{b.status}</span>
              </td>
              <td>{new Date(b.created_at).toLocaleString('en-IN')}</td>
              <td>
                {b.status === 'active' ? (
                  <button
                    className="btn btn-danger btn-sm"
                    onClick={() => handleCancel(b.id)}
                    disabled={cancellingId === b.id}
                  >
                    {cancellingId === b.id ? 'Cancelling...' : 'Cancel'}
                  </button>
                ) : (
                  '—'
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      </div>
    </div>
  )
}
