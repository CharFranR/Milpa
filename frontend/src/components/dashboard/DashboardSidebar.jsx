import { useLocation, useNavigate } from 'react-router-dom'
import Icon from '../ui/Icon'
import Avatar from '../Avatar'
import { useAuth } from '../../context/AuthContext'
import { getInitials, getDisplayName } from '../../lib/user'
import { BUYER_TAB_PATHS } from '../../lib/routes'
import { cn } from '../../lib/cn'
import { ROLE_LABELS } from '../../lib/roles'

const MAYORISTA_TABS = [
  { id: 'solicitudes', icon: 'view_list', label: 'Mis solicitudes' },
  { id: 'mensajes', icon: 'chat_bubble', label: 'Mensajes' },
  { id: 'perfil', icon: 'person', label: 'Mi perfil' },
  { id: 'marketplace', icon: 'storefront', label: 'Marketplace' },
]

const MINORISTA_TABS = [
  { id: 'mensajes', icon: 'chat_bubble', label: 'Mensajes' },
  { id: 'perfil', icon: 'person', label: 'Mi perfil' },
  { id: 'marketplace', icon: 'storefront', label: 'Marketplace' },
]

function TabButton({ tab, active, onSelect, className }) {
  return (
    <button
      type="button"
      onClick={() => onSelect(tab.id)}
      aria-current={active ? 'page' : undefined}
      className={cn(
        'inline-flex shrink-0 items-center gap-2.5 rounded-xl px-4 py-2.5 text-sm font-medium transition-colors',
        'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand',
        active ? 'bg-brand text-white' : 'text-gray-600 hover:bg-brand-soft hover:text-brand',
        className,
      )}
    >
      <Icon name={tab.icon} size={19} />
      {tab.label}
    </button>
  )
}

export default function DashboardSidebar() {
  const navigate = useNavigate()
  const { pathname } = useLocation()
  const { user, logout } = useAuth()
  const displayName = getDisplayName(user)
  const rawRole = user?.role
  const displayRole = rawRole ? ROLE_LABELS[rawRole] : 'Comprador'

  const tabs = rawRole === 'mayorista' ? MAYORISTA_TABS : MINORISTA_TABS

  function selectTab(id) {
    navigate(BUYER_TAB_PATHS[id])
  }

  function handleLogout() {
    navigate('/')
    logout()
  }

  return (
    <>
      <aside className="hidden w-64 shrink-0 lg:block">
        <div className="sticky top-20 rounded-2xl border border-gray-100 bg-white p-4">
          <div className="flex items-center gap-3 px-2 py-2">
            <Avatar initials={getInitials(user)} name={displayName} size="md" />
            <div className="min-w-0">
              <p className="truncate text-sm font-bold text-gray-900">{displayName}</p>
              <p className="truncate text-xs text-gray-500">
                {displayRole}
              </p>
            </div>
          </div>

          <nav aria-label="Secciones del dashboard" className="mt-4 space-y-1">
            {tabs.map((tab) => (
              <TabButton
                key={tab.id}
                tab={tab}
                active={pathname === BUYER_TAB_PATHS[tab.id]}
                onSelect={selectTab}
                className="w-full"
              />
            ))}
          </nav>

          <div className="my-4 border-t border-gray-100" />

          <button
            type="button"
            onClick={handleLogout}
            className="mt-1 inline-flex w-full items-center gap-2.5 rounded-xl px-4 py-2.5 text-sm font-medium text-red-600 transition-colors hover:bg-red-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-red-600"
          >
            <Icon name="logout" size={19} />
            Cerrar sesión
          </button>
        </div>
      </aside>

      <div className="sticky top-16 z-30 border-b border-gray-100 bg-white/95 backdrop-blur lg:hidden">
        <div className="flex items-center gap-2 overflow-x-auto px-4 py-3">
          <Avatar initials={getInitials(user)} name={displayName} size="sm" />
          {tabs.map((tab) => (
            <TabButton
              key={tab.id}
              tab={tab}
              active={pathname === BUYER_TAB_PATHS[tab.id]}
              onSelect={selectTab}
              className="px-3 py-2 text-xs"
            />
          ))}
          <button
            type="button"
            onClick={handleLogout}
            aria-label="Cerrar sesión"
            className="ml-auto inline-flex shrink-0 rounded-xl p-2 text-red-600 transition-colors hover:bg-red-50"
          >
            <Icon name="logout" size={20} />
          </button>
        </div>
      </div>
    </>
  )
}