import Icon from '../components/ui/Icon'
import MarketplaceCatalog from '../components/marketplace/MarketplaceCatalog'
import { Link } from 'react-router-dom'

export default function Marketplace() {
  return (
    <main className="mx-auto w-full max-w-7xl flex-1 px-4 py-8 sm:px-6 lg:px-8">
      <nav aria-label="Ruta de navegación" className="flex items-center gap-1 text-sm text-gray-500">
        <Link to="/" className="hover:text-brand">
          Inicio
        </Link>
        <Icon name="chevron_right" size={16} className="text-gray-300" />
        <span className="font-semibold text-gray-900">Marketplace</span>
      </nav>

      <MarketplaceCatalog />
    </main>
  )
}
