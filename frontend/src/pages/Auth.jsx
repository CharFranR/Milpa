import { useEffect, useState } from 'react'
import Login from '../components/auth/Login'
import Register from '../components/auth/Register'

export default function Auth() {
  const [view, setView] = useState(() => viewFromHash(window.location.hash))

  useEffect(() => {
    const update = () => setView(viewFromHash(window.location.hash))
    window.addEventListener('hashchange', update)
    return () => window.removeEventListener('hashchange', update)
  }, [])

  return view === 'register' ? <Register /> : <Login />
}

function viewFromHash(hash) {
  return hash === '#/register' ? 'register' : 'login'
}
