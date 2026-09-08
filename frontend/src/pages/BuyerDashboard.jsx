import { useState } from 'react'
import Navbar from '../components/layout/Navbar'
import DashboardSidebar from '../components/dashboard/DashboardSidebar'
import MarketplaceCatalog from '../components/marketplace/MarketplaceCatalog'
import BuyerHome from './buyer/BuyerHome'
import BuyerFavorites from './buyer/BuyerFavorites'
import BuyerMessages from './buyer/BuyerMessages'
import BuyerProfile from './buyer/BuyerProfile'

export default function BuyerDashboard() {
  const [tab, setTab] = useState('inicio')

  return (
    <div className="flex min-h-screen flex-col bg-gray-50">
      <Navbar />

      <div className="mx-auto flex w-full max-w-7xl flex-1 flex-col px-4 py-8 sm:px-6 lg:flex-row lg:gap-8 lg:px-8">
        <DashboardSidebar activeTab={tab} onTabChange={setTab} />

        <main className="mt-6 min-w-0 flex-1 lg:mt-0">
          {tab === 'inicio' && <BuyerHome onGoToTab={setTab} />}
          {tab === 'favoritos' && <BuyerFavorites />}
          {tab === 'mensajes' && <BuyerMessages />}
          {tab === 'perfil' && <BuyerProfile />}
          {tab === 'marketplace' && <MarketplaceCatalog />}
        </main>
      </div>
    </div>
  )
}
