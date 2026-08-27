import { cn } from '../../lib/cn'
import Icon from '../ui/Icon'
import Button from '../ui/Button'
import { clearSessionRole } from '../../lib/session'

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

export default function AdminSidebar({ activeTab, onTabChange }) {
  function handleLogout() {
    clearSessionRole()
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

  return (
    <aside className="hidden lg:block w-64 shrink-0">
      <div className="flex flex-col h-full">
        <div className="p-4 border-b border-gray-100">
          <div className="flex items-center gap-3">
            <span
              className="flex h-12 w-12 items-center justify-center rounded-2xl bg-night text-white relative"
              role="img"
              aria-label="Admin Principal"
            >
              <span className="text-xl font-bold">AD</span>
              <span className="absolute bottom-0 right-0 h-3 w-3 rounded-full bg-green-500 ring-2 ring-white" aria-label="En línea" />
            </span>
            <div className="min-w-0">
              <p className="font-semibold text-gray-900 truncate">Admin Principal</p>
              <p className="text-xs text-gray-500 truncate">Administrador</p>
            </div>
          </div>
        </div>

        <nav className="flex-1 p-3 space-y-1" aria-label="Gestión">
          <p className="px-3 py-1 text-xs font-semibold uppercase tracking-wider text-gray-400">GESTIÓN</p>
          {GESTION_TABS.map((tab) => (
            <button
              key={tab.id}
              type="button"
              onClick={() => onTabChange(tab.id)}
              className={cn(
                'w-full flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors',
                activeTab === tab.id
                  ? 'bg-brand-soft text-brand'
                  : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900',
              )}
            >
              <Icon name={tab.icon} size={19} />
              {tab.label}
              <Badge count={tab.badge} />
            </button>
          ))}
        </nav>

        <nav className="px-3 pb-3" aria-label="Sistema">
          <p className="px-3 py-1 text-xs font-semibold uppercase tracking-wider text-gray-400">SISTEMA</p>
          <div className="space-y-1">
            {SISTEMA_TABS.map((tab) => (
              <button
                key={tab.id}
                type="button"
                onClick={() => onTabChange(tab.id)}
                className={cn(
                  'w-full flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors',
                  activeTab === tab.id
                    ? 'bg-brand-soft text-brand'
                    : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900',
                )}
              >
                <Icon name={tab.icon} size={19} />
                {tab.label}
              </button>
            ))}
          </div>
        </nav>

        <div className="p-3 border-t border-gray-100">
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
      </div>
    </aside>
  )
}