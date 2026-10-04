import { Route, Routes } from 'react-router-dom'
import RequireRole from './components/auth/RequireRole'
import Login from './components/auth/Login'
import Register from './components/auth/Register'
import MarketplaceCatalog from './components/marketplace/MarketplaceCatalog'
import AdminConfig from './pages/admin/AdminConfig'
import AdminHome from './pages/admin/AdminHome'
import AdminModeration from './pages/admin/AdminModeration'
import AdminProducts from './pages/admin/AdminProducts'
import AdminProducers from './pages/admin/AdminProducers'
import AdminReports from './pages/admin/AdminReports'
import AdminUsers from './pages/admin/AdminUsers'
import BuyerFavorites from './pages/buyer/BuyerFavorites'
import BuyerHome from './pages/buyer/BuyerHome'
import BuyerMessages from './pages/buyer/BuyerMessages'
import BuyerProfile from './pages/buyer/BuyerProfile'
import Landing from './pages/Landing'
import Marketplace from './pages/Marketplace'
import ProductDetail from './pages/ProductDetail'
import ProducerBusiness from './pages/producer/ProducerBusiness'
import ProducerHome from './pages/producer/ProducerHome'
import ProducerMessages from './pages/producer/ProducerMessages'
import ProducerProducts from './pages/producer/ProducerProducts'
import ProducerRequests from './pages/producer/ProducerRequests'
import AdminDashboard from './pages/AdminDashboard'
import BuyerDashboard from './pages/BuyerDashboard'
import ProducerDashboard from './pages/ProducerDashboard'

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<Landing />} />
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />

      <Route
        path="/marketplace"
        element={
          <RequireRole role="buyer">
            <Marketplace />
          </RequireRole>
        }
      />
      <Route
        path="/product/:id"
        element={
          <RequireRole role="buyer">
            <ProductDetail />
          </RequireRole>
        }
      />

      <Route
        path="/dashboard"
        element={
          <RequireRole role="buyer">
            <BuyerDashboard />
          </RequireRole>
        }
      >
        <Route index element={<BuyerHome />} />
        <Route path="favorites" element={<BuyerFavorites />} />
        <Route path="messages" element={<BuyerMessages />} />
        <Route path="profile" element={<BuyerProfile />} />
        <Route path="marketplace" element={<MarketplaceCatalog />} />
      </Route>

      <Route
        path="/producer"
        element={
          <RequireRole role="producer">
            <ProducerDashboard />
          </RequireRole>
        }
      >
        <Route index element={<ProducerHome />} />
        <Route path="products" element={<ProducerProducts />} />
        <Route path="requests" element={<ProducerRequests />} />
        <Route path="messages" element={<ProducerMessages />} />
        <Route path="business" element={<ProducerBusiness />} />
      </Route>

      <Route
        path="/admin"
        element={
          <RequireRole role="admin">
            <AdminDashboard />
          </RequireRole>
        }
      >
        <Route index element={<AdminHome />} />
        <Route path="users" element={<AdminUsers />} />
        <Route path="producers" element={<AdminProducers />} />
        <Route path="products" element={<AdminProducts />} />
        <Route path="moderation" element={<AdminModeration />} />
        <Route path="reports" element={<AdminReports />} />
        <Route path="settings" element={<AdminConfig />} />
      </Route>

      <Route path="*" element={<Landing />} />
    </Routes>
  )
}
