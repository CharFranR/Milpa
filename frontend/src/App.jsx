import { Suspense, lazy } from 'react'
import { Route, Routes } from 'react-router-dom'
import RequireRole from './components/auth/RequireRole'
import Spinner from './components/ui/Spinner'

const Login = lazy(() => import('./components/auth/Login'))
const Register = lazy(() => import('./components/auth/Register'))
const MarketplaceCatalog = lazy(() => import('./components/marketplace/MarketplaceCatalog'))
const AdminConfig = lazy(() => import('./pages/admin/AdminConfig'))
const AdminHome = lazy(() => import('./pages/admin/AdminHome'))
const AdminModeration = lazy(() => import('./pages/admin/AdminModeration'))
const AdminProducts = lazy(() => import('./pages/admin/AdminProducts'))
const AdminProducers = lazy(() => import('./pages/admin/AdminProducers'))
const AdminReports = lazy(() => import('./pages/admin/AdminReports'))
const AdminUsers = lazy(() => import('./pages/admin/AdminUsers'))
const BuyerMessages = lazy(() => import('./pages/buyer/BuyerMessages'))
const BuyerProfile = lazy(() => import('./pages/buyer/BuyerProfile'))
const BuyerRequestDetail = lazy(() => import('./pages/buyer/BuyerRequestDetail'))
const BuyerRequestForm = lazy(() => import('./pages/buyer/BuyerRequestForm'))
const BuyerRequests = lazy(() => import('./pages/buyer/BuyerRequests'))
const BuyerHome = lazy(() => import('./pages/buyer/BuyerHome'))
const Landing = lazy(() => import('./pages/Landing'))
const Liquidations = lazy(() => import('./pages/Liquidations'))
const MatchDetail = lazy(() => import('./pages/matches/MatchDetail'))
const Marketplace = lazy(() => import('./pages/Marketplace'))
const ProductDetail = lazy(() => import('./pages/ProductDetail'))
const ProducerBusiness = lazy(() => import('./pages/producer/ProducerBusiness'))
const ProducerHome = lazy(() => import('./pages/producer/ProducerHome'))
const ProducerMessages = lazy(() => import('./pages/producer/ProducerMessages'))
const ProducerInventory = lazy(() => import('./pages/producer/ProducerInventory'))
const ProducerOfferForm = lazy(() => import('./pages/producer/ProducerOfferForm'))
const ProducerLiquidations = lazy(() => import('./pages/producer/ProducerLiquidations'))
const ProducerOffers = lazy(() => import('./pages/producer/ProducerOffers'))
const ProducerProducts = lazy(() => import('./pages/producer/ProducerProducts'))
const ProducerRequests = lazy(() => import('./pages/producer/ProducerRequests'))
const ProducerAvailable = lazy(() => import('./pages/producer/ProducerAvailable'))
const AdminDashboard = lazy(() => import('./pages/AdminDashboard'))
const BuyerDashboard = lazy(() => import('./pages/BuyerDashboard'))
const ProducerDashboard = lazy(() => import('./pages/ProducerDashboard'))

function RouteFallback() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <Spinner size={32} />
    </div>
  )
}

export default function App() {
  const buyerRoles = ['minorista', 'mayorista']

  return (
    <Suspense fallback={<RouteFallback />}>
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
    </Suspense>
  )
}
