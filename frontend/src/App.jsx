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
import BuyerMessages from './pages/buyer/BuyerMessages'
import BuyerProfile from './pages/buyer/BuyerProfile'
import BuyerRequestDetail from './pages/buyer/BuyerRequestDetail'
import BuyerRequestForm from './pages/buyer/BuyerRequestForm'
import BuyerRequests from './pages/buyer/BuyerRequests'
import BuyerHome from './pages/buyer/BuyerHome'
import Landing from './pages/Landing'
import Liquidations from './pages/Liquidations'
import MatchDetail from './pages/matches/MatchDetail'
import Marketplace from './pages/Marketplace'
import ProductDetail from './pages/ProductDetail'
import ProducerBusiness from './pages/producer/ProducerBusiness'
import ProducerHome from './pages/producer/ProducerHome'
import ProducerMessages from './pages/producer/ProducerMessages'
import ProducerInventory from './pages/producer/ProducerInventory'
import ProducerOfferForm from './pages/producer/ProducerOfferForm'
import ProducerLiquidations from './pages/producer/ProducerLiquidations'
import ProducerOffers from './pages/producer/ProducerOffers'
import ProducerProducts from './pages/producer/ProducerProducts'
import ProducerRequests from './pages/producer/ProducerRequests'
import ProducerAvailable from './pages/producer/ProducerAvailable'
import AdminDashboard from './pages/AdminDashboard'
import BuyerDashboard from './pages/BuyerDashboard'
import ProducerDashboard from './pages/ProducerDashboard'

export default function App() {
  const buyerRoles = ['minorista', 'mayorista']

  return (
    <Routes>
      <Route path="/" element={<Landing />} />
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />
      <Route path="/liquidations" element={<Liquidations />} />

      <Route
        path="/marketplace"
        element={
          <RequireRole roles={buyerRoles}>
            <Marketplace />
          </RequireRole>
        }
      />
      <Route
        path="/product/:id"
        element={
          <RequireRole roles={buyerRoles}>
            <ProductDetail />
          </RequireRole>
        }
      />

      <Route
        path="/dashboard"
        element={
          <RequireRole roles={buyerRoles}>
            <BuyerDashboard />
          </RequireRole>
        }
      >
        <Route
          index
          element={
            <RequireRole roles={buyerRoles}>
              <BuyerHome />
            </RequireRole>
          }
        />
        <Route
          path="requests"
          element={
            <RequireRole roles={['mayorista']}>
              <BuyerRequests />
            </RequireRole>
          }
        />
        <Route
          path="requests/new"
          element={
            <RequireRole roles={['mayorista']}>
              <BuyerRequestForm />
            </RequireRole>
          }
        />
        <Route
          path="requests/:id"
          element={
            <RequireRole roles={['mayorista']}>
              <BuyerRequestDetail />
            </RequireRole>
          }
        />
        <Route
          path="requests/:id/edit"
          element={
            <RequireRole roles={['mayorista']}>
              <BuyerRequestForm />
            </RequireRole>
          }
        />
        <Route
          path="matches/:matchId"
          element={
            <RequireRole roles={['mayorista']}>
              <MatchDetail />
            </RequireRole>
          }
        />
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
        <Route path="available" element={<ProducerAvailable />} />
        <Route path="available/:id/offer" element={<ProducerOfferForm />} />
        <Route path="offers" element={<ProducerOffers />} />
        <Route path="matches/:matchId" element={<MatchDetail />} />
        <Route path="inventory" element={<ProducerInventory />} />
        <Route path="products" element={<ProducerProducts />} />
        <Route path="liquidations" element={<ProducerLiquidations />} />
        <Route path="requests" element={<ProducerRequests />} />
        <Route path="messages" element={<ProducerMessages />} />
        <Route path="business" element={<ProducerBusiness />} />
      </Route>

      <Route
        path="/admin"
        element={
          <RequireRole roles={['admin', 'auditor']}>
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