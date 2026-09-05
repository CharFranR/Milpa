import Icon from '../ui/Icon'
import Avatar from '../Avatar'
import { getUser, clearSession } from '../../lib/session'
import { getInitials, getDisplayName } from '../../lib/user'
import { cn } from '../../lib/cn'

const TABS = [
  { id: 'inicio', icon: 'home', label: 'Inicio' },
  { id: 'favoritos', icon: 'favorite', label: 'Favoritos' },
  { id: 'mensajes', icon: 'chat_bubble', label: 'Mensajes' },
  { id: 'perfil', icon: 'person', label: 'Mi perfil' },
  { id: 'marketplace', icon: 'storefront', label: 'Marketplace' },
]

function handleLogout() {
  clearSession()
  window.location.hash = '#/'
}

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

export default function DashboardSidebar({ activeTab, onTabChange }) {
  const user = getUser()
  const displayName = getDisplayName(user)
  const displayRole = user?.role === 'buyer' ? 'Comprador' : user?.role || 'Comprador'

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
            {TABS.map((tab) => (
              <TabButton
                key={tab.id}
                tab={tab}
                active={activeTab === tab.id}
                onSelect={onTabChange}
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
          {TABS.map((tab) => (
            <TabButton
              key={tab.id}
              tab={tab}
              active={activeTab === tab.id}
              onSelect={onTabChange}
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
