import { useState } from 'react'
import Navbar from '../components/layout/Navbar'
import AdminSidebar from '../components/dashboard/AdminSidebar'
import AdminHome from './admin/AdminHome'
import AdminUsers from './admin/AdminUsers'
import AdminProducers from './admin/AdminProducers'
import AdminProducts from './admin/AdminProducts'
import AdminModeration from './admin/AdminModeration'
import AdminReports from './admin/AdminReports'

export default function AdminDashboard() {
  const [tab, setTab] = useState('dashboard')

  return (
    <div className="flex min-h-screen flex-col bg-gray-50">
      <Navbar />

      <div className="mx-auto flex w-full max-w-7xl flex-1 flex-col px-4 py-8 sm:px-6 lg:flex-row lg:gap-8 lg:px-8">
        <AdminSidebar activeTab={tab} onTabChange={setTab} />

        <main className="mt-6 min-w-0 flex-1 lg:mt-0">
          {tab === 'dashboard' && <AdminHome />}
          {tab === 'usuarios' && <AdminUsers />}
          {tab === 'productores' && <AdminProducers />}
          {tab === 'productos' && <AdminProducts />}
          {tab === 'moderacion' && <AdminModeration />}
          {tab === 'reportes' && <AdminReports />}
          {tab === 'configuracion' && <AdminConfig />}
        </main>
      </div>
    </div>
  )
}

function AdminConfig() {
  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold text-gray-900">Configuración</h1>
      <div className="rounded-xl border border-gray-100 bg-white p-6">
        <p className="text-gray-500">Panel de configuración del sistema (sin implementar).</p>
      </div>
    </div>
  )
}