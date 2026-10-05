import { Navigate, useLocation } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'

export default function RequireRole({ role, roles, children }) {
  const location = useLocation()
  const { role: sessionRole } = useAuth()

  const allowed = roles || [role]
  if (!allowed.includes(sessionRole)) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />
  }

  return children
}
