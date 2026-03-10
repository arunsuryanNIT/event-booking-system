import { useUser } from '../context/UserContext'

/**
 * UserPicker renders a dropdown to switch the currently selected user.
 * Lives inside the Navbar so it's always accessible.
 */
export default function UserPicker() {
  const { currentUser, setCurrentUser, users } = useUser()

  const handleChange = (e) => {
    const selected = users.find((u) => u.id === e.target.value)
    if (selected) setCurrentUser(selected)
  }

  if (!currentUser) return null

  return (
    <select
      className="user-picker"
      value={currentUser.id}
      onChange={handleChange}
    >
      {users.map((u) => (
        <option key={u.id} value={u.id}>
          {u.name}
        </option>
      ))}
    </select>
  )
}
