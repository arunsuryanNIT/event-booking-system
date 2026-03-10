import { useState, useEffect, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import { getEvent } from '../api/client'
import BookingButton from '../components/BookingButton'

/** EventPage shows full event details and a booking button. */
export default function EventPage() {
  const { id } = useParams()
  const [event, setEvent] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  const fetchEvent = useCallback((showLoading = true) => {
    if (showLoading) setLoading(true)
    getEvent(id)
      .then((res) => setEvent(res.data))
      .catch((err) => setError(err.error || 'Failed to load event'))
      .finally(() => setLoading(false))
  }, [id])

  useEffect(() => {
    fetchEvent()
  }, [fetchEvent])

  if (loading) return <p>Loading...</p>
  if (error) return <p className="msg msg-error">{error}</p>
  if (!event) return null

  const remaining = event.capacity - event.booked_count

  return (
    <div>
      <Link to="/" className="back-link">← Back to Events</Link>
      <div className="event-detail">
        <h2>{event.title}</h2>
        <p className="event-meta">
          {new Date(event.event_date).toLocaleDateString('en-IN', {
            weekday: 'long', day: 'numeric', month: 'long', year: 'numeric',
            hour: '2-digit', minute: '2-digit',
          })}
          {' · '}
          {event.location}
        </p>
        <p>{event.description}</p>
        <div className="event-stats">
          <span>Capacity: {event.capacity}</span>
          <span>Booked: {event.booked_count}</span>
          <span>Remaining: {remaining}</span>
        </div>
        <BookingButton event={event} onBooked={() => fetchEvent(false)} />
      </div>
    </div>
  )
}
