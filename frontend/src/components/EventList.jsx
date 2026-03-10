import EventCard from './EventCard'

/**
 * EventList renders a grid of EventCards.
 *
 * @param {{ events: Array }} props
 */
export default function EventList({ events }) {
  if (!events.length) {
    return <p className="empty-state">No events found.</p>
  }

  return (
    <div className="event-grid">
      {events.map((event) => (
        <EventCard key={event.id} event={event} />
      ))}
    </div>
  )
}
