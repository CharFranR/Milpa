import { cn } from '../../lib/cn'
import Icon from '../ui/Icon'
import Button from '../ui/Button'
import { clearSession } from '../../lib/session'
import { producerProfile } from '../../mocks/producer'

const TABS = [
  { id: 'resumen', icon: 'home', label: 'Resumen' },
  { id: 'productos', icon: 'inventory_2', label: 'Mis productos' },
  { id: 'solicitudes', icon: 'inbox', label: 'Solicitudes' },
  { id: 'mensajes', icon: 'chat_bubble', label: 'Mensajes' },
  { id: 'negocio', icon: 'storefront', label: 'Mi negocio' },
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

export default function ProducerSidebar({ activeTab, onTabChange }) {
  return (
    <>
      <aside className="hidden w-64 shrink-0 lg:block">
        <div className="sticky top-20 rounded-2xl border border-gray-100 bg-white p-4">
          <div className="flex items-center gap-3 px-2 py-2">
            <span
              className="flex h-12 w-12 items-center justify-center rounded-2xl bg-brand text-white"
              role="img"
              aria-label="María González"
            >
              <Icon name="agriculture" size={24} />
            </span>
            <div className="min-w-0">
              <p className="truncate text-sm font-bold text-gray-900">{producerProfile.name}</p>
              <p className="truncate text-xs text-gray-500">{producerProfile.farm}</p>
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

          <div className="space-y-2">
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
            <button
              type="button"
              onClick={handleLogout}
              className="mt-1 inline-flex w-full items-center gap-2.5 rounded-xl px-4 py-2.5 text-sm font-medium text-red-600 transition-colors hover:bg-red-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-red-600"
            >
              <Icon name="logout" size={19} />
              Cerrar sesión
            </button>
          </div>
        </div>
      </aside>

      <div className="sticky top-16 z-30 border-b border-gray-100 bg-white/95 backdrop-blur lg:hidden">
        <div className="flex items-center gap-2 overflow-x-auto px-4 py-3">
          <span
            className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-2xl bg-brand text-white"
            role="img"
            aria-label="María González"
          >
            <Icon name="agriculture" size={18} />
          </span>
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