import { useState } from 'react'
import Icon from '../../components/ui/Icon'
import Button from '../../components/ui/Button'
import { useUserProfile } from '../../hooks/useUserProfile'
import { getUser } from '../../lib/session'

const SECURITY_ITEMS = [
  { icon: 'lock', label: 'Cambiar contraseña' },
  { icon: 'security', label: 'Verificación en dos pasos' },
  { icon: 'devices', label: 'Sesiones activas' },
]

export default function BuyerProfile() {
  const currentUser = getUser()
  const { user, loading, error, updateUser } = useUserProfile(currentUser?.id)
  const [editing, setEditing] = useState(false)
  const [saving, setSaving] = useState(false)
  const [form, setForm] = useState({ first_name: '', last_name: '', email: '', phone_number: '', address: '' })
  const [success, setSuccess] = useState('')

  function startEdit() {
    setForm({
      first_name: user?.first_name || '',
      last_name: user?.last_name || '',
      email: user?.email || '',
      phone_number: user?.phone_number || '',
      address: user?.address || '',
    })
    setEditing(true)
    setSuccess('')
  }

  function setField(key, value) {
    setForm((f) => ({ ...f, [key]: value }))
  }

  function handleSave(e) {
    e.preventDefault()
    setSaving(true)
    updateUser(form)
      .then(() => {
        setEditing(false)
        setSuccess('Perfil actualizado correctamente.')
      })
      .catch(() => {})
      .finally(() => setSaving(false))
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <header>
          <h1 className="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">Mi perfil</h1>
          <p className="mt-1 text-sm text-gray-500">Administra tu información personal y seguridad.</p>
        </header>
        <div className="grid gap-6 xl:grid-cols-2">
          <div className="animate-pulse rounded-2xl border border-gray-100 bg-white p-6 space-y-4">
            <div className="h-5 w-32 rounded bg-gray-200" />
            <div className="h-10 rounded bg-gray-200" />
            <div className="h-10 rounded bg-gray-200" />
            <div className="h-10 rounded bg-gray-200" />
          </div>
          <div className="animate-pulse rounded-2xl border border-gray-100 bg-white p-6 space-y-4">
            <div className="h-5 w-24 rounded bg-gray-200" />
            <div className="h-14 rounded bg-gray-200" />
            <div className="h-14 rounded bg-gray-200" />
          </div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-6">
        <header>
          <h1 className="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">Mi perfil</h1>
        </header>
        <div className="rounded-2xl border border-red-200 bg-red-50 p-6 text-center">
          <Icon name="error" size={40} className="mx-auto text-red-400" />
          <p className="mt-3 text-sm text-red-700">{error}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <header>
        <h1 className="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">Mi perfil</h1>
        <p className="mt-1 text-sm text-gray-500">Administra tu información personal y seguridad.</p>
      </header>

      {success && (
        <p className="rounded-lg bg-green-50 px-3 py-2 text-sm text-green-700">{success}</p>
      )}

      <div className="grid gap-6 xl:grid-cols-2">
        <section
          aria-label="Datos personales"
          className="flex flex-col rounded-2xl border border-gray-100 bg-white p-6"
        >
          <h2 className="text-base font-bold text-gray-900">Datos personales</h2>

          {editing ? (
            <form onSubmit={handleSave} className="mt-4 flex-1 space-y-4">
              <div>
                <label className="text-xs font-semibold text-gray-600">Nombre</label>
                <input
                  type="text"
                  required
                  value={form.first_name}
                  onChange={(e) => setField('first_name', e.target.value)}
                  className="mt-1.5 w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
                />
              </div>
              <div>
                <label className="text-xs font-semibold text-gray-600">Apellido</label>
                <input
                  type="text"
                  required
                  value={form.last_name}
                  onChange={(e) => setField('last_name', e.target.value)}
                  className="mt-1.5 w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
                />
              </div>
              <div>
                <label className="text-xs font-semibold text-gray-600">Correo electrónico</label>
                <input
                  type="email"
                  required
                  value={form.email}
                  onChange={(e) => setField('email', e.target.value)}
                  className="mt-1.5 w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
                />
              </div>
              <div>
                <label className="text-xs font-semibold text-gray-600">Teléfono</label>
                <input
                  type="tel"
                  value={form.phone_number}
                  onChange={(e) => setField('phone_number', e.target.value)}
                  className="mt-1.5 w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
                />
              </div>
              <div>
                <label className="text-xs font-semibold text-gray-600">Dirección</label>
                <input
                  type="text"
                  value={form.address}
                  onChange={(e) => setField('address', e.target.value)}
                  className="mt-1.5 w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
                />
              </div>
              <div className="flex gap-3 pt-2">
                <Button type="button" variant="outline" onClick={() => setEditing(false)} disabled={saving}>
                  Cancelar
                </Button>
                <Button type="submit" variant="primary" disabled={saving}>
                  {saving ? (
                    <>
                      <Icon name="progress_activity" size={16} className="animate-spin" />
                      Guardando...
                    </>
                  ) : 'Guardar cambios'}
                </Button>
              </div>
            </form>
          ) : (
            <dl className="mt-4 flex-1 space-y-4">
              <div>
                <dt className="text-xs font-semibold text-gray-600">Nombre completo</dt>
                <dd className="mt-1 rounded-lg bg-gray-50 px-3 py-2.5 text-sm text-gray-900">
                  {user?.first_name} {user?.last_name}
                </dd>
              </div>
              <div>
                <dt className="text-xs font-semibold text-gray-600">Correo electrónico</dt>
                <dd className="mt-1 rounded-lg bg-gray-50 px-3 py-2.5 text-sm text-gray-900">
                  {user?.email}
                </dd>
              </div>
              <div>
                <dt className="text-xs font-semibold text-gray-600">Teléfono</dt>
                <dd className="mt-1 rounded-lg bg-gray-50 px-3 py-2.5 text-sm text-gray-900">
                  {user?.phone_number || 'No registrado'}
                </dd>
              </div>
              <div>
                <dt className="text-xs font-semibold text-gray-600">Dirección</dt>
                <dd className="mt-1 rounded-lg bg-gray-50 px-3 py-2.5 text-sm text-gray-900">
                  {user?.address || 'No registrada'}
                </dd>
              </div>
              <Button type="button" variant="outline" className="mt-5 w-full" onClick={startEdit} icon={<Icon name="edit" size={16} />}>
                Editar información
              </Button>
            </dl>
          )}
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
