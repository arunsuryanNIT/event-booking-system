import { Link } from 'react-router-dom'

/**
 * EventCard displays a single event summary: title, date, location,
 * capacity info, and a link to the detail page.
 *
 * @param {{ event: Object }} props
 */
export default function EventCard({ event }) {
  const remaining = event.capacity - event.booked_count
  const pct = (event.booked_count / event.capacity) * 100

  return (
    <div className="event-card">
      <h3>{event.title}</h3>
      <p className="event-meta">
        {new Date(event.event_date).toLocaleDateString('en-IN', {
          day: 'numeric', month: 'short', year: 'numeric',
          hour: '2-digit', minute: '2-digit',
        })}
        {' · '}
        {event.location}
      </p>
      <p className="event-desc">{event.description}</p>
      <div className="capacity-bar">
        <div className="capacity-fill" style={{ width: `${pct}%` }} />
      </div>
      <p className="capacity-text">
        {remaining > 0 ? `${remaining} / ${event.capacity} spots left` : 'Sold Out'}
      </p>
      <Link to={`/events/${event.id}`} className="btn btn-outline">
        View Details
      </Link>
    </div>
  )
}
