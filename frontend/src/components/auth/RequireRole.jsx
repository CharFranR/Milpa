import { Navigate, useLocation } from 'react-router-dom'
import { hasRole } from '../../lib/session'

export default function RequireRole({ role, children }) {
  const location = useLocation()

  if (!hasRole(role)) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />
  }

  return children
}
