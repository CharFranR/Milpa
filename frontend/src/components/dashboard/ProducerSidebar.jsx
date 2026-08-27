import { cn } from '../../lib/cn'
import Icon from '../ui/Icon'
import Button from '../ui/Button'
import { clearSessionRole } from '../../lib/session'

const TABS = [
  { id: 'resumen', icon: 'home', label: 'Resumen' },
  { id: 'productos', icon: 'inventory_2', label: 'Mis productos' },
  { id: 'solicitudes', icon: 'inbox', label: 'Solicitudes' },
  { id: 'mensajes', icon: 'chat_bubble', label: 'Mensajes' },
  { id: 'negocio', icon: 'storefront', label: 'Mi negocio' },
]

export default function ProducerSidebar({ activeTab, onTabChange }) {
  function handleLogout() {
    clearSessionRole()
    window.location.hash = '#/'
  }

  return (
    <aside className="hidden lg:block w-64 shrink-0">
      <div className="flex flex-col h-full">
        <div className="p-4 border-b border-gray-100">
          <div className="flex items-center gap-3">
            <span
              className="flex h-12 w-12 items-center justify-center rounded-2xl bg-brand text-white"
              role="img"
              aria-label="María González"
            >
              <Icon name="agriculture" size={24} />
            </span>
            <div className="min-w-0">
              <p className="font-semibold text-gray-900 truncate">María González</p>
              <p className="text-xs text-gray-500 truncate">Finca La Esperanza</p>
            </div>
          </div>
        </div>

        <nav className="flex-1 p-3 space-y-1" aria-label="Navegación del dashboard">
          {TABS.map((tab) => (
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
        </nav>

        <div className="p-3 space-y-2 border-t border-gray-100">
          <Button
            type="button"
            variant="primary"
            size="sm"
            className="w-full"
            onClick={() => onTabChange('productos')}
            icon={<Icon name="add" size={16} />}
          >
            Agregar producto
          </Button>
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