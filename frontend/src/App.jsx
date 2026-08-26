import { useEffect, useState } from 'react'
import Landing from './pages/Landing'
import Auth from './pages/Auth'
import Marketplace from './pages/Marketplace'
import { hasBuyerSession } from './lib/session'

export default function App() {
  const [route, setRoute] = useState(() => routeFromHash(window.location.hash))

  useEffect(() => {
    const onHashChange = () => setRoute(routeFromHash(window.location.hash))
    window.addEventListener('hashchange', onHashChange)
    return () => window.removeEventListener('hashchange', onHashChange)
  }, [])

  useEffect(() => {
    if (route === 'marketplace-forbidden') {
      window.location.hash = '#/login'
    }
  }, [route])

  if (route === 'auth' || route === 'marketplace-forbidden') return <Auth />
  if (route === 'marketplace') return <Marketplace />
  return <Landing />
}

function routeFromHash(hash) {
  if (hash === '#/login' || hash === '#/register') return 'auth'
  if (hash === '#/marketplace') {
    return hasBuyerSession() ? 'marketplace' : 'marketplace-forbidden'
  }
  return 'landing'
}
