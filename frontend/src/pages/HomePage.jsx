import { useState, useEffect } from 'react'
import { getEvents } from '../api/client'
import EventList from '../components/EventList'

/** HomePage fetches and displays all events. */
export default function HomePage() {
  const [events, setEvents] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    getEvents()
      .then((res) => setEvents(res.data || []))
      .finally(() => setLoading(false))
  }, [])

  return (
    <div>
      <h2>Upcoming Events</h2>
      {loading ? <p>Loading...</p> : <EventList events={events} />}
    </div>
  )
}
