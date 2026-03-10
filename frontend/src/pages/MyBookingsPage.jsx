import { useState, useEffect, useCallback } from 'react'
import { getUserBookings } from '../api/client'
import { useUser } from '../context/UserContext'
import MyBookings from '../components/MyBookings'

/** MyBookingsPage fetches and displays the current user's bookings. */
export default function MyBookingsPage() {
  const { currentUser } = useUser()
  const [bookings, setBookings] = useState([])
  const [loading, setLoading] = useState(true)

  const fetchBookings = useCallback(() => {
    if (!currentUser) return
    setLoading(true)
    getUserBookings(currentUser.id)
      .then((res) => setBookings(res.data || []))
      .finally(() => setLoading(false))
  }, [currentUser])

  useEffect(() => {
    fetchBookings()
  }, [fetchBookings])

  if (!currentUser) return <p>Select a user to view bookings.</p>

  return (
    <div>
      <h2>My Bookings — {currentUser.name}</h2>
      {loading ? <p>Loading...</p> : (
        <MyBookings bookings={bookings} onCancelled={fetchBookings} />
      )}
    </div>
  )
}
