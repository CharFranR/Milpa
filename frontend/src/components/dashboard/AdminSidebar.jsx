import { cn } from '../../lib/cn'
import Icon from '../ui/Icon'
import Button from '../ui/Button'
import { clearSession } from '../../lib/session'

const GESTION_TABS = [
  { id: 'dashboard', icon: 'dashboard', label: 'Dashboard', badge: null },
  { id: 'usuarios', icon: 'group', label: 'Usuarios', badge: 3 },
  { id: 'productores', icon: 'agriculture', label: 'Productores', badge: null },
  { id: 'productos', icon: 'inventory_2', label: 'Productos', badge: 7 },
  { id: 'moderacion', icon: 'shield', label: 'Moderación', badge: 2 },
  { id: 'reportes', icon: 'analytics', label: 'Reportes', badge: null },
]

const SISTEMA_TABS = [
  { id: 'configuracion', icon: 'settings', label: 'Configuración', badge: null },
]

const ALL_TABS = [...GESTION_TABS, ...SISTEMA_TABS]

function handleLogout() {
  clearSession()
  window.location.hash = '#/'
}

function Badge({ count }) {
  if (!count) return null
  return (
    <span className="ml-auto inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-red-500 px-1.5 text-xs font-bold text-white">
      {count}
    </span>
  )
}

function TabButton({ tab, active, onSelect, className, showBadge }) {
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
      {showBadge && <Badge count={tab.badge} />}
    </button>
  )
}

export default function AdminSidebar({ activeTab, onTabChange }) {
  return (
    <>
      <aside className="hidden w-64 shrink-0 lg:block">
        <div className="sticky top-20 rounded-2xl border border-gray-100 bg-white p-4">
          <div className="flex items-center gap-3 px-2 py-2">
            <span
              className="flex h-12 w-12 items-center justify-center rounded-2xl bg-night text-white relative"
              role="img"
              aria-label="Admin Principal"
            >
              <span className="text-xl font-bold">AD</span>
              <span className="absolute bottom-0 right-0 h-3 w-3 rounded-full bg-green-500 ring-2 ring-white" aria-label="En línea" />
            </span>
            <div className="min-w-0">
              <p className="truncate text-sm font-bold text-gray-900">Admin Principal</p>
              <p className="truncate text-xs text-gray-500">Administrador</p>
            </div>
          </div>

          <nav aria-label="Gestión" className="mt-4 space-y-1">
            <p className="px-2 py-1 text-xs font-semibold uppercase tracking-wider text-gray-400">GESTIÓN</p>
            {GESTION_TABS.map((tab) => (
              <TabButton
                key={tab.id}
                tab={tab}
                active={activeTab === tab.id}
                onSelect={onTabChange}
                className="w-full"
                showBadge
              />
            ))}
          </nav>

          <nav aria-label="Sistema" className="mt-4 space-y-1">
            <p className="px-2 py-1 text-xs font-semibold uppercase tracking-wider text-gray-400">SISTEMA</p>
            {SISTEMA_TABS.map((tab) => (
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

          <Button
            type="button"
            variant="outline"
            size="sm"
            className="w-full text-red-600 hover:bg-red-50"
            onClick={handleLogout}
            icon={<Icon name="logout" size={16} />}
          >
            Cerrar sesión
          </Button>
        </div>
      </aside>

      <div className="sticky top-16 z-30 border-b border-gray-100 bg-white/95 backdrop-blur lg:hidden">
        <div className="flex items-center gap-2 overflow-x-auto px-4 py-3">
          <span
            className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-2xl bg-night text-white relative"
            role="img"
            aria-label="Admin Principal"
          >
            <span className="text-xs font-bold">AD</span>
            <span className="absolute bottom-0 right-0 h-2.5 w-2.5 rounded-full bg-green-500 ring-2 ring-white" />
          </span>
          {ALL_TABS.map((tab) => (
            <TabButton
              key={tab.id}
              tab={tab}
              active={activeTab === tab.id}
              onSelect={onTabChange}
              className="px-3 py-2 text-xs"
              showBadge
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