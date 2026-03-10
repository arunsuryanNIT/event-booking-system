import { createContext, useContext, useState, useEffect } from 'react'
import { getUsers } from '../api/client'

const UserContext = createContext()

/**
 * UserProvider fetches all pre-seeded users on mount and exposes
 * the currently selected user via React Context. The first user
 * is auto-selected so the app is immediately usable.
 */
export function UserProvider({ children }) {
  const [currentUser, setCurrentUser] = useState(null)
  const [users, setUsers] = useState([])

  useEffect(() => {
    getUsers().then((res) => {
      setUsers(res.data)
      if (res.data.length > 0) {
        setCurrentUser(res.data[0])
      }
    })
  }, [])

  return (
    <UserContext.Provider value={{ currentUser, setCurrentUser, users }}>
      {children}
    </UserContext.Provider>
  )
}

/** @returns {{ currentUser: Object|null, setCurrentUser: Function, users: Array }} */
export function useUser() {
  return useContext(UserContext)
}
