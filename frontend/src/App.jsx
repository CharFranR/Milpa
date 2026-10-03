import { useEffect, useState } from 'react'
import Landing from './pages/Landing'
import Auth from './pages/Auth'
import Marketplace from './pages/Marketplace'
import BuyerDashboard from './pages/BuyerDashboard'
import ProducerDashboard from './pages/ProducerDashboard'
import AdminDashboard from './pages/AdminDashboard'
import ProductDetail from './pages/ProductDetail'
import { hasRole } from './lib/session'

export default function App() {
  const [route, setRoute] = useState(() => resolveRoute(window.location.hash))

  useEffect(() => {
    const onHashChange = () => setRoute(resolveRoute(window.location.hash))
    window.addEventListener('hashchange', onHashChange)
    return () => window.removeEventListener('hashchange', onHashChange)
  }, [])

  useEffect(() => {
    const hash = window.location.hash
    if (route === 'auth' && (hash === '#/marketplace' || hash === '#/dashboard' || hash === '#/producer' || hash === '#/admin' || hash.startsWith('#/product/'))) {
      window.location.hash = '#/login'
    }
  }, [route])

  if (route === 'auth') return <Auth />
  if (route === 'marketplace') return <Marketplace />
  if (route === 'dashboard') return <BuyerDashboard />
  if (route === 'producer') return <ProducerDashboard />
  if (route === 'admin') return <AdminDashboard />
  if (route === 'product') return <ProductDetail />
  return <Landing />
}

function resolveRoute(hash) {
  if (hash === '#/login' || hash === '#/register') return 'auth'
  if (hash === '#/marketplace') return hasRole('buyer') ? 'marketplace' : 'auth'
  if (hash === '#/dashboard') return hasRole('buyer') ? 'dashboard' : 'auth'
  if (hash.startsWith('#/product/')) return hasRole('buyer') ? 'product' : 'auth'
  if (hash === '#/producer') return hasRole('producer') ? 'producer' : 'auth'
  if (hash === '#/admin') return 'admin'
  return 'landing'
}
