import ProducerSidebar from '../components/dashboard/ProducerSidebar'
import DashboardLayout from '../components/layout/DashboardLayout'

export default function ProducerDashboard() {
  return <DashboardLayout sidebar={<ProducerSidebar />} />
}
