import { useEffect, useState } from 'react'
import Landing from './pages/Landing'
import Auth from './pages/Auth'
import Marketplace from './pages/Marketplace'
import BuyerDashboard from './pages/BuyerDashboard'
import ProducerDashboard from './pages/ProducerDashboard'
import AdminDashboard from './pages/AdminDashboard'
import { hasBuyerSession, hasProducerSession, hasAdminSession } from './lib/session'

export default function App() {
  const [route, setRoute] = useState(() => resolveRoute(window.location.hash))

  useEffect(() => {
    const onHashChange = () => setRoute(resolveRoute(window.location.hash))
    window.addEventListener('hashchange', onHashChange)
    return () => window.removeEventListener('hashchange', onHashChange)
  }, [])

  useEffect(() => {
    const hash = window.location.hash
    if (route === 'auth' && (hash === '#/marketplace' || hash === '#/dashboard' || hash === '#/producer' || hash === '#/admin')) {
      window.location.hash = '#/login'
    }
  }, [route])

  if (route === 'auth') return <Auth />
  if (route === 'marketplace') return <Marketplace />
  if (route === 'dashboard') return <BuyerDashboard />
  if (route === 'producer') return <ProducerDashboard />
  if (route === 'admin') return <AdminDashboard />
  return <Landing />
}

function resolveRoute(hash) {
  if (hash === '#/login' || hash === '#/register') return 'auth'
  if (hash === '#/marketplace') {
    return hasBuyerSession() ? 'marketplace' : 'auth'
  }
  if (hash === '#/dashboard') {
    return hasBuyerSession() ? 'dashboard' : 'auth'
  }
  if (hash === '#/producer') {
    return hasProducerSession() ? 'producer' : 'auth'
  }
  if (hash === '#/admin') {
    return hasAdminSession() ? 'admin' : 'auth'
  }
  return 'landing'
}
