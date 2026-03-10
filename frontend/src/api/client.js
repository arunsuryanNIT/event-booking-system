/**
 * API client using the native fetch API.
 * All functions return parsed JSON with the standard API envelope:
 * { success, data, error, message }
 *
 * On non-2xx responses the parsed error body is thrown,
 * so callers can catch and read error.message or error.error.
 */

const BASE = '/api'

async function api(path, options = {}) {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  const data = await res.json()
  if (!res.ok) throw data
  return data
}

/** @returns {Promise<{data: Array}>} all pre-seeded users */
export const getUsers = () => api('/users')

/** @returns {Promise<{data: Array}>} all events ordered by date */
export const getEvents = () => api('/events')

/**
 * @param {string} id - event UUID
 * @returns {Promise<{data: Object}>} single event
 */
export const getEvent = (id) => api(`/events/${id}`)

/**
 * @param {string} eventId - event UUID
 * @param {string} userId - user UUID
 * @returns {Promise<{data: Object}>} created booking
 */
export const bookEvent = (eventId, userId) =>
  api(`/events/${eventId}/book`, {
    method: 'POST',
    body: JSON.stringify({ user_id: userId }),
  })

/**
 * @param {string} bookingId - booking UUID
 * @param {string} userId - user UUID
 * @returns {Promise<{data: Object}>} cancelled booking
 */
export const cancelBooking = (bookingId, userId) =>
  api(`/bookings/${bookingId}/cancel`, {
    method: 'POST',
    body: JSON.stringify({ user_id: userId }),
  })

/**
 * @param {string} userId - user UUID
 * @returns {Promise<{data: Array}>} user's bookings with event titles
 */
export const getUserBookings = (userId) =>
  api(`/users/${userId}/bookings`)

/**
 * @param {Object} filters - optional: event_id, user_id, booking_id, operation, outcome
 * @returns {Promise<{data: Array}>} filtered audit log entries
 */
export const getAuditLogs = (filters = {}) => {
  const params = new URLSearchParams()
  Object.entries(filters).forEach(([key, val]) => {
    if (val) params.append(key, val)
  })
  const qs = params.toString()
  return api(`/audit${qs ? `?${qs}` : ''}`)
}
