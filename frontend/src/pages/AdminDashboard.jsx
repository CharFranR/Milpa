import AdminSidebar from '../components/dashboard/AdminSidebar'
import DashboardLayout from '../components/layout/DashboardLayout'

export default function AdminDashboard() {
  return <DashboardLayout sidebar={<AdminSidebar />} />
}
