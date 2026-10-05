import { useAuth } from '../../context/AuthContext'
import BuyerRequests from './BuyerRequests'
import MarketplaceCatalog from '../../components/marketplace/MarketplaceCatalog'

export default function BuyerHome() {
  const { role } = useAuth()

  if (role === 'mayorista') {
    return <BuyerRequests />
  }

  return <MarketplaceCatalog />
}