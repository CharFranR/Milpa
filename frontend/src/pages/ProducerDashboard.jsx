import { useState } from 'react'
import Navbar from '../components/layout/Navbar'
import ProducerSidebar from '../components/dashboard/ProducerSidebar'
import ProducerHome from './producer/ProducerHome'
import ProducerProducts from './producer/ProducerProducts'
import ProducerRequests from './producer/ProducerRequests'
import ProducerMessages from './producer/ProducerMessages'
import ProducerBusiness from './producer/ProducerBusiness'

export default function ProducerDashboard() {
  const [tab, setTab] = useState('resumen')

  return (
    <div className="flex min-h-screen flex-col bg-gray-50">
      <Navbar />

      <div className="mx-auto flex w-full max-w-7xl flex-1 flex-col px-4 py-8 sm:px-6 lg:flex-row lg:gap-8 lg:px-8">
        <ProducerSidebar activeTab={tab} onTabChange={setTab} />

        <main className="mt-6 min-w-0 flex-1 lg:mt-0">
          {tab === 'resumen' && <ProducerHome />}
          {tab === 'productos' && <ProducerProducts />}
          {tab === 'solicitudes' && <ProducerRequests />}
          {tab === 'mensajes' && <ProducerMessages />}
          {tab === 'negocio' && <ProducerBusiness />}
        </main>
      </div>
    </div>
  )
}