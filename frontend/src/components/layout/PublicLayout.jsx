import { Outlet, useLocation } from 'react-router-dom'
import { cn } from '../../lib/cn'
import Navbar from './Navbar'
import Footer from './Footer'

// La landing se dibuja sin fondo; el resto de páginas públicas, sobre gris.
const GRAY_ROUTES = ['/marketplace']

export default function PublicLayout() {
  const { pathname } = useLocation()
  const gray = pathname.startsWith('/product/') || GRAY_ROUTES.includes(pathname)

  return (
    <div className={cn('flex min-h-screen flex-col', gray && 'bg-gray-50')}>
      <Navbar />
      <Outlet />
      <Footer />
    </div>
  )
}
