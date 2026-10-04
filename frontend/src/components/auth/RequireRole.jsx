import { Navigate, useLocation } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'

export default function RequireRole({ role, children }) {
  const location = useLocation()
  const { role: sessionRole } = useAuth()

  if (sessionRole !== role) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />
  }

  return children
}
