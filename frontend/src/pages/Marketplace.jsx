import Navbar from '../components/layout/Navbar'
import Footer from '../components/layout/Footer'
import Icon from '../components/ui/Icon'
import MarketplaceCatalog from '../components/marketplace/MarketplaceCatalog'

export default function Marketplace() {
  return (
    <div className="flex min-h-screen flex-col bg-gray-50">
      <Navbar />

      <main className="mx-auto w-full max-w-7xl flex-1 px-4 py-8 sm:px-6 lg:px-8">
        <nav aria-label="Ruta de navegación" className="flex items-center gap-1 text-sm text-gray-500">
          <a href="#/" className="hover:text-brand">
            Inicio
          </a>
          <Icon name="chevron_right" size={16} className="text-gray-300" />
          <span className="font-semibold text-gray-900">Marketplace</span>
        </nav>

        <MarketplaceCatalog />
      </main>

      <Footer />
    </div>
  )
}
