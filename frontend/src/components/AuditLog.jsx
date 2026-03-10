import { useState, useEffect } from 'react'
import { getAuditLogs, getEvents, getUsers } from '../api/client'

/**
 * AuditLog renders the audit log table with filter dropdowns.
 * Filters trigger an immediate refetch via useEffect.
 */
export default function AuditLog() {
  const [logs, setLogs] = useState([])
  const [events, setEvents] = useState([])
  const [users, setUsers] = useState([])
  const [filters, setFilters] = useState({
    event_id: '',
    user_id: '',
    operation: '',
    outcome: '',
  })

  useEffect(() => {
    getEvents().then((res) => setEvents(res.data || []))
    getUsers().then((res) => setUsers(res.data || []))
  }, [])

  useEffect(() => {
    getAuditLogs(filters).then((res) => setLogs(res.data || []))
  }, [filters])

  const updateFilter = (key, value) => {
    setFilters((prev) => ({ ...prev, [key]: value }))
  }

  const clearFilters = () => {
    setFilters({ event_id: '', user_id: '', operation: '', outcome: '' })
  }

  return (
    <div>
      <div className="filters">
        <select value={filters.event_id} onChange={(e) => updateFilter('event_id', e.target.value)}>
          <option value="">All Events</option>
          {events.map((ev) => (
            <option key={ev.id} value={ev.id}>{ev.title}</option>
          ))}
        </select>

        <select value={filters.user_id} onChange={(e) => updateFilter('user_id', e.target.value)}>
          <option value="">All Users</option>
          {users.map((u) => (
            <option key={u.id} value={u.id}>{u.name}</option>
          ))}
        </select>

        <select value={filters.operation} onChange={(e) => updateFilter('operation', e.target.value)}>
          <option value="">All Operations</option>
          <option value="book">Book</option>
          <option value="cancel">Cancel</option>
        </select>

        <select value={filters.outcome} onChange={(e) => updateFilter('outcome', e.target.value)}>
          <option value="">All Outcomes</option>
          <option value="success">Success</option>
          <option value="failure">Failure</option>
        </select>

        <button className="btn btn-outline btn-sm" onClick={clearFilters}>
          Clear Filters
        </button>
      </div>

      {logs.length === 0 ? (
        <p className="empty-state">No audit log entries found.</p>
      ) : (
        <div className="table-wrap">
        <table className="table">
          <thead>
            <tr>
              <th>Timestamp</th>
              <th>Operation</th>
              <th>Event</th>
              <th>User</th>
              <th>Booking ID</th>
              <th>Outcome</th>
              <th>Reason</th>
            </tr>
          </thead>
          <tbody>
            {logs.map((log) => (
              <tr key={log.id}>
                <td>{new Date(log.created_at).toLocaleString('en-IN')}</td>
                <td>
                  <span className={`badge badge-${log.operation}`}>
                    {log.operation}
                  </span>
                </td>
                <td>{log.event_title}</td>
                <td>{log.user_name}</td>
                <td>{log.booking_id || '—'}</td>
                <td>
                  {log.outcome === 'success'
                    ? <span className="outcome-success">✓</span>
                    : <span className="outcome-failure">✗</span>}
                </td>
                <td>{log.failure_reason || '—'}</td>
              </tr>
            ))}
          </tbody>
        </table>
        </div>
      )}
    </div>
  )
}
