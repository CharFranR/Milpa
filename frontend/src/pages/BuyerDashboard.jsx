import { Outlet } from 'react-router-dom'
import Navbar from '../components/layout/Navbar'
import DashboardSidebar from '../components/dashboard/DashboardSidebar'

export default function BuyerDashboard() {
  return (
    <div className="flex min-h-screen flex-col bg-gray-50">
      <Navbar />

      <div className="mx-auto flex w-full max-w-7xl flex-1 flex-col px-4 py-8 sm:px-6 lg:flex-row lg:gap-8 lg:px-8">
        <DashboardSidebar />

        <main className="mt-6 min-w-0 flex-1 lg:mt-0">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
