import { useEffect, useState } from 'react'
import Landing from './pages/Landing'
import Auth from './pages/Auth'
import Marketplace from './pages/Marketplace'

const ROUTES = {
  '#/login': 'auth',
  '#/register': 'auth',
  '#/marketplace': 'marketplace',
}

export default function App() {
  const [route, setRoute] = useState(() => routeFromHash(window.location.hash))

  useEffect(() => {
    const onHashChange = () => setRoute(routeFromHash(window.location.hash))
    window.addEventListener('hashchange', onHashChange)
    return () => window.removeEventListener('hashchange', onHashChange)
  }, [])

  if (route === 'auth') return <Auth />
  if (route === 'marketplace') return <Marketplace />
  return <Landing />
}

function routeFromHash(hash) {
  return ROUTES[hash] ?? 'landing'
}
