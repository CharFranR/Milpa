import { useEffect, useState } from 'react'
import Landing from './pages/Landing'
import Auth from './pages/Auth'

export default function App() {
  const [atAuth, setAtAuth] = useState(() => isAuthRoute(window.location.hash))

  useEffect(() => {
    const onHashChange = () => setAtAuth(isAuthRoute(window.location.hash))
    window.addEventListener('hashchange', onHashChange)
    return () => window.removeEventListener('hashchange', onHashChange)
  }, [])

  return atAuth ? <Auth /> : <Landing />
}

function isAuthRoute(hash) {
  return hash === '#/login' || hash === '#/register'
}
