import Icon from '../../components/ui/Icon'
import Button from '../../components/ui/Button'
import { buyerProfile } from '../../mocks/buyer'

const SECURITY_ITEMS = [
  { icon: 'lock', label: 'Cambiar contraseña' },
  { icon: 'security', label: 'Verificación en dos pasos' },
  { icon: 'devices', label: 'Sesiones activas' },
]

export default function BuyerProfile() {
  return (
    <div className="space-y-6">
      <header>
        <h1 className="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">Mi perfil</h1>
        <p className="mt-1 text-sm text-gray-500">Administra tu información personal y seguridad.</p>
      </header>

      <div className="grid gap-6 xl:grid-cols-2">
        <section
          aria-label="Datos personales"
          className="flex flex-col rounded-2xl border border-gray-100 bg-white p-6"
        >
          <h2 className="text-base font-bold text-gray-900">Datos personales</h2>

          <dl className="mt-4 flex-1 space-y-4">
            <div>
              <dt className="text-xs font-semibold text-gray-600">Nombre completo</dt>
              <dd className="mt-1 rounded-lg bg-gray-50 px-3 py-2.5 text-sm text-gray-900">
                {buyerProfile.fullName}
              </dd>
            </div>
            <div>
              <dt className="text-xs font-semibold text-gray-600">Correo electrónico</dt>
              <dd className="mt-1 rounded-lg bg-gray-50 px-3 py-2.5 text-sm text-gray-900">
                {buyerProfile.email}
              </dd>
            </div>
            <div>
              <dt className="text-xs font-semibold text-gray-600">Teléfono</dt>
              <dd className="mt-1 rounded-lg bg-gray-50 px-3 py-2.5 text-sm text-gray-900">
                {buyerProfile.phone}
              </dd>
            </div>
            <div>
              <dt className="text-xs font-semibold text-gray-600">Ciudad</dt>
              <dd className="mt-1 rounded-lg bg-gray-50 px-3 py-2.5 text-sm text-gray-900">
                {buyerProfile.city}, {buyerProfile.region}
              </dd>
            </div>
          </dl>

          <Button type="button" variant="outline" className="mt-5 w-full" icon={<Icon name="edit" size={16} />}>
            Editar información
          </Button>
        </section>

        <section aria-label="Seguridad" className="rounded-2xl border border-gray-100 bg-white p-6">
          <h2 className="text-base font-bold text-gray-900">Seguridad</h2>

          <ul role="list" className="mt-4 divide-y divide-gray-100">
            {SECURITY_ITEMS.map((item) => (
              <li key={item.label}>
                <button
                  type="button"
                  className="flex w-full items-center gap-3 py-4 text-left transition-colors hover:text-brand focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand"
                >
                  <span className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-brand-soft text-brand">
                    <Icon name={item.icon} size={19} />
                  </span>
                  <span className="flex-1 text-sm font-medium text-gray-800">{item.label}</span>
                  <Icon name="chevron_right" size={18} className="text-gray-300" />
                </button>
              </li>
            ))}
          </ul>
        </section>
      </div>
    </div>
  )
}
