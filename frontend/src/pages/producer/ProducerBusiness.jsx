import { useState, useEffect } from 'react'
import Icon from '../../components/ui/Icon'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'
import { useCompany } from '../../hooks/useCompany'
import { categories } from '../../services/api'
import { getUser } from '../../lib/session'

export default function ProducerBusiness() {
  const currentUser = getUser()
  const { company, loading, error, createCompany, updateCompany } = useCompany(currentUser?.id)
  const [editing, setEditing] = useState(false)
  const [saving, setSaving] = useState(false)
  const [formError, setFormError] = useState('')
  const [success, setSuccess] = useState('')
  const [cats, setCats] = useState([])
  const [form, setForm] = useState({
    name: '',
    category_id: '',
    address: '',
    description: '',
    phone_number: '',
    email: '',
    website: '',
  })

  useEffect(() => {
    categories.getAll().then((data) => setCats(Array.isArray(data) ? data : [])).catch(() => {})
  }, [])

  function setField(key, value) {
    setForm((f) => ({ ...f, [key]: value }))
  }

  function startCreate() {
    setForm({
      name: '',
      category_id: '',
      address: '',
      description: '',
      phone_number: currentUser?.phone_number || '',
      email: currentUser?.email || '',
      website: '',
    })
    setEditing(true)
    setFormError('')
  }

  function startEdit() {
    setForm({
      name: company?.name || '',
      category_id: company?.category_id || '',
      address: company?.address || '',
      description: company?.description || '',
      phone_number: company?.phone_number || '',
      email: company?.email || '',
      website: company?.website || '',
    })
    setEditing(true)
    setFormError('')
    setSuccess('')
  }

  function handleSave(e) {
    e.preventDefault()
    if (!form.name.trim()) {
      setFormError('El nombre de la empresa es obligatorio.')
      return
    }
    setSaving(true)
    setFormError('')

    const payload = { ...form }
    if (!payload.category_id) {
      delete payload.category_id
    }

    const action = company
      ? updateCompany(company.id, payload)
      : createCompany(payload)

    action
      .then(() => {
        setEditing(false)
        setSuccess(company ? 'Empresa actualizada.' : 'Empresa creada correctamente.')
      })
      .catch((err) => {
        setFormError(err.message || 'Error al guardar.')
      })
      .finally(() => setSaving(false))
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <h1 className="text-3xl font-bold text-gray-900">Mi negocio</h1>
        <div className="grid gap-6 lg:grid-cols-2">
          <div className="animate-pulse rounded-xl border border-gray-100 bg-white p-6 space-y-4">
            <div className="flex items-center gap-4">
              <div className="h-16 w-16 rounded-full bg-gray-200" />
              <div className="space-y-2 flex-1">
                <div className="h-5 w-48 rounded bg-gray-200" />
                <div className="h-4 w-32 rounded bg-gray-200" />
              </div>
            </div>
            <div className="space-y-3">
              <div className="h-4 rounded bg-gray-200" />
              <div className="h-4 rounded bg-gray-200" />
            </div>
          </div>
          <div className="animate-pulse rounded-xl border border-gray-100 bg-white p-6 space-y-4">
            <div className="h-5 w-32 rounded bg-gray-200" />
            <div className="flex gap-2">
              <div className="h-8 w-24 rounded-full bg-gray-200" />
              <div className="h-8 w-20 rounded-full bg-gray-200" />
            </div>
          </div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-6">
        <h1 className="text-3xl font-bold text-gray-900">Mi negocio</h1>
        <div className="rounded-xl border border-red-200 bg-red-50 p-6 text-center">
          <Icon name="error" size={40} className="mx-auto text-red-400" />
          <p className="mt-3 text-sm text-red-700">{error}</p>
        </div>
      </div>
    )
  }

  if (!company && !editing) {
    return (
      <div className="space-y-6">
        <h1 className="text-3xl font-bold text-gray-900">Mi negocio</h1>
        <div className="rounded-xl border border-dashed border-gray-300 bg-white p-12 text-center">
          <span className="flex h-16 w-16 mx-auto items-center justify-center rounded-2xl bg-brand-soft text-brand">
            <Icon name="storefront" size={32} />
          </span>
          <h2 className="mt-4 text-lg font-semibold text-gray-900">Aún no has creado tu empresa</h2>
          <p className="mt-2 max-w-sm mx-auto text-sm text-gray-500">
            Crea tu perfil de empresa para publicar productos y aparecer en el Marketplace.
          </p>
          <Button type="button" variant="primary" className="mt-6" onClick={startCreate} icon={<Icon name="add" size={18} />}>
            Crear mi empresa
          </Button>
        </div>
      </div>
    )
  }

  if (editing) {
    return (
      <div className="space-y-6">
        <h1 className="text-3xl font-bold text-gray-900">
          {company ? 'Editar empresa' : 'Crear empresa'}
        </h1>

        {formError && (
          <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{formError}</p>
        )}

        <form onSubmit={handleSave} className="rounded-xl border border-gray-200 bg-white p-6 space-y-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="sm:col-span-2">
              <label className="text-xs font-semibold text-gray-600">Nombre de la empresa *</label>
              <input
                type="text"
                required
                value={form.name}
                onChange={(e) => setField('name', e.target.value)}
                placeholder="Mi Empresa Agrícola"
                className="mt-1.5 w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
              />
            </div>
            <div>
              <label className="text-xs font-semibold text-gray-600">Categoría principal</label>
              <select
                value={form.category_id}
                onChange={(e) => setField('category_id', e.target.value)}
                className="mt-1.5 w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
              >
                <option value="">Sin categoría</option>
                {cats.map((c) => (
                  <option key={c.id} value={c.id}>{c.name}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="text-xs font-semibold text-gray-600">Dirección</label>
              <input
                type="text"
                value={form.address}
                onChange={(e) => setField('address', e.target.value)}
                placeholder="Managua, Nicaragua"
                className="mt-1.5 w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
              />
            </div>
            <div>
              <label className="text-xs font-semibold text-gray-600">Teléfono</label>
              <input
                type="tel"
                value={form.phone_number}
                onChange={(e) => setField('phone_number', e.target.value)}
                placeholder="+505 8888 1234"
                className="mt-1.5 w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
              />
            </div>
            <div>
              <label className="text-xs font-semibold text-gray-600">Correo electrónico</label>
              <input
                type="email"
                value={form.email}
                onChange={(e) => setField('email', e.target.value)}
                placeholder="empresa@correo.com"
                className="mt-1.5 w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
              />
            </div>
            <div className="sm:col-span-2">
              <label className="text-xs font-semibold text-gray-600">Sitio web</label>
              <input
                type="url"
                value={form.website}
                onChange={(e) => setField('website', e.target.value)}
                placeholder="https://miempresa.com"
                className="mt-1.5 w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
              />
            </div>
            <div className="sm:col-span-2">
              <label className="text-xs font-semibold text-gray-600">Descripción</label>
              <textarea
                rows={3}
                value={form.description}
                onChange={(e) => setField('description', e.target.value)}
                placeholder="Describe tu empresa, especialidades, horarios, redes sociales..."
                className="mt-1.5 w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand resize-none"
              />
            </div>
          </div>
          <div className="flex justify-end gap-3 pt-2">
            <Button type="button" variant="outline" onClick={() => setEditing(false)} disabled={saving}>
              Cancelar
            </Button>
            <Button type="submit" variant="primary" disabled={saving}>
              {saving ? (
                <>
                  <Icon name="progress_activity" size={16} className="animate-spin" />
                  Guardando...
                </>
              ) : company ? 'Guardar cambios' : 'Crear empresa'}
            </Button>
          </div>
        </form>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold text-gray-900">Mi negocio</h1>

      {success && (
        <p className="rounded-lg bg-green-50 px-3 py-2 text-sm text-green-700">{success}</p>
      )}

      <div className="grid gap-6 lg:grid-cols-2">
        <section aria-label="Información de la empresa" className="rounded-xl border border-gray-100 bg-white p-6">
          <div className="flex items-start gap-4">
            <span className="flex h-16 w-16 shrink-0 items-center justify-center rounded-2xl bg-brand text-white">
              <Icon name="storefront" size={24} />
            </span>
            <div className="min-w-0 flex-1">
              <h2 className="text-xl font-bold text-gray-900">{company.name}</h2>
              <p className="mt-0.5 text-gray-500">{company.address || 'Sin dirección'}</p>
            </div>
          </div>
          <dl className="mt-6 space-y-4">
            <div className="flex items-center gap-3">
              <Icon name="mail" size={18} className="text-gray-400 shrink-0" />
              <div>
                <dt className="text-xs font-semibold text-gray-500">Correo</dt>
                <dd className="text-gray-900">{company.email || 'No registrado'}</dd>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <Icon name="phone" size={18} className="text-gray-400 shrink-0" />
              <div>
                <dt className="text-xs font-semibold text-gray-500">Teléfono</dt>
                <dd className="text-gray-900">{company.phone_number || 'No registrado'}</dd>
              </div>
            </div>
            {company.website && (
              <div className="flex items-center gap-3">
                <Icon name="language" size={18} className="text-gray-400 shrink-0" />
                <div>
                  <dt className="text-xs font-semibold text-gray-500">Sitio web</dt>
                  <dd className="text-gray-900">{company.website}</dd>
                </div>
              </div>
            )}
            {company.description && (
              <div className="flex items-start gap-3">
                <Icon name="info" size={18} className="text-gray-400 shrink-0 mt-0.5" />
                <div>
                  <dt className="text-xs font-semibold text-gray-500">Descripción</dt>
                  <dd className="text-gray-900 text-sm leading-relaxed">{company.description}</dd>
                </div>
              </div>
            )}
          </dl>
          <Button type="button" variant="outline" className="mt-6 w-full" onClick={startEdit} icon={<Icon name="edit" size={16} />}>
            Editar empresa
          </Button>
        </section>

        <section aria-label="Categoría" className="rounded-xl border border-gray-100 bg-white p-6">
          <h2 className="text-lg font-semibold text-gray-900">Categoría</h2>
          <div className="mt-3">
            {company.category_id ? (
              <Badge tone="brand">{company.category_id}</Badge>
            ) : (
              <p className="text-sm text-gray-500">Sin categoría asignada</p>
            )}
          </div>
        </section>
      </div>
    </div>
  )
}
