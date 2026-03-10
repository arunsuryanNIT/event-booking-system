import { useState, useEffect } from 'react'
import { bookEvent, getUserBookings } from '../api/client'
import { useUser } from '../context/UserContext'

/**
 * BookingButton handles the booking flow for a single event.
 * On mount it checks if the current user already has an active booking.
 * States: Book Now → Booking... → Already Booked / Sold Out.
 * On success, a centered confirmation popup appears with a close button.
 * Behind the popup the button immediately shows "Already Booked".
 *
 * @param {{ event: Object, onBooked: Function }} props
 */
export default function BookingButton({ event, onBooked }) {
  const { currentUser } = useUser()
  const [loading, setLoading] = useState(false)
  const [alreadyBooked, setAlreadyBooked] = useState(false)
  const [showPopup, setShowPopup] = useState(false)
  const [error, setError] = useState(null)

  useEffect(() => {
    if (!currentUser) return
    getUserBookings(currentUser.id).then((res) => {
      const hasActive = (res.data || []).some(
        (b) => b.event_id === event.id && b.status === 'active'
      )
      setAlreadyBooked(hasActive)
    })
  }, [currentUser, event.id])

  const remaining = event.capacity - event.booked_count

  const handleBook = async () => {
    if (!currentUser) return
    setLoading(true)
    setError(null)
    try {
      await bookEvent(event.id, currentUser.id)
      setAlreadyBooked(true)
      setShowPopup(true);
      if (onBooked) onBooked()
    } catch (err) {
      const msg = err.error || 'Booking failed'
      if (msg.includes('already')) {
        setAlreadyBooked(true)
      }
      setError(msg)
    } finally {
      setLoading(false)
    }
  }

  const buttonLabel = () => {
    if (alreadyBooked) return 'Already Booked'
    if (remaining <= 0) return 'Sold Out'
    if (loading) return 'Booking...'
    return 'Book Now'
  }

  const isDisabled = alreadyBooked || remaining <= 0 || loading

  return (
    <div>
      <button
        className={`btn ${isDisabled ? 'btn-disabled' : 'btn-primary'}`}
        onClick={handleBook}
        disabled={isDisabled}
      >
        {buttonLabel()}
      </button>
   
      {error && <p className="msg msg-error">{error}</p>}

      {showPopup && (
        <div className="popup-overlay" onClick={() => setShowPopup(false)}>
          <div className="popup" onClick={(e) => e.stopPropagation()}>
            <button className="popup-close" onClick={() => setShowPopup(false)}>
              ✕
            </button>
            <div className="popup-icon">✓</div>
            <h3>Booking Confirmed</h3>
            <p>Your booking has been confirmed.</p>
          </div>
        </div>
      )}
    </div>
  )
}
