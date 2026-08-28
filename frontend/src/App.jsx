import { useEffect, useState } from 'react'
import Landing from './pages/Landing'
import Auth from './pages/Auth'
import Marketplace from './pages/Marketplace'
import BuyerDashboard from './pages/BuyerDashboard'
import { hasBuyerSession } from './lib/session'

export default function App() {
  const [route, setRoute] = useState(() => resolveRoute(window.location.hash))

  useEffect(() => {
    const onHashChange = () => setRoute(resolveRoute(window.location.hash))
    window.addEventListener('hashchange', onHashChange)
    return () => window.removeEventListener('hashchange', onHashChange)
  }, [])

  useEffect(() => {
    const hash = window.location.hash
    if (route === 'auth' && (hash === '#/marketplace' || hash === '#/dashboard')) {
      window.location.hash = '#/login'
    }
  }, [route])

  if (route === 'auth') return <Auth />
  if (route === 'marketplace') return <Marketplace />
  if (route === 'dashboard') return <BuyerDashboard />
  return <Landing />
}

function resolveRoute(hash) {
  if (hash === '#/login' || hash === '#/register') return 'auth'
  if (hash === '#/marketplace' || hash === '#/dashboard') {
    return hasBuyerSession() ? (hash === '#/dashboard' ? 'dashboard' : 'marketplace') : 'auth'
  }
  return 'landing'
}
