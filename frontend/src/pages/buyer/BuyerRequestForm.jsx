import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import Button from '../../components/ui/Button'
import ErrorState from '../../components/ui/ErrorState'
import Icon from '../../components/ui/Icon'
import Skeleton from '../../components/ui/Skeleton'
import { useSupplyRequest } from '../../hooks/useSupplyRequests'
import { supplyRequests } from '../../services/supplyRequests'
import { MEASUREMENT_LABELS } from '../../lib/supplyStatus'
import { EMPTY_FORM, buildPayload, toForm, validate } from '../../lib/supplyRequestForm'

const INPUT =
  'mt-1.5 w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand'
const LABEL = 'text-xs font-semibold text-gray-600'

export default function BuyerRequestForm() {
  const { id } = useParams()
  const editing = Boolean(id)
  const navigate = useNavigate()
  const { request, loading, error: loadError, reload } = useSupplyRequest(id)

  const [form, setForm] = useState(EMPTY_FORM)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (request) setForm(toForm(request))
  }, [request])

  function setField(field, value) {
    setForm((prev) => ({ ...prev, [field]: value }))
  }

  async function handleSubmit(event) {
    event.preventDefault()
    const validationError = validate(form)
    if (validationError) {
      setError(validationError)
      return
    }

    setSaving(true)
    setError('')
    try {
      if (editing) {
        await supplyRequests.update(id, buildPayload(form, request, true))
        navigate(`/dashboard/requests/${id}`)
      } else {
        const created = await supplyRequests.create(buildPayload(form, null, false))
        navigate(`/dashboard/requests/${created.id}`)
      }
    } catch (err) {
      setError(err.message || 'No se pudo guardar la solicitud.')
      setSaving(false)
    }
  }

  if (editing && loading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-64" />
        <div className="rounded-2xl border border-gray-100 bg-white p-6">
          <Skeleton className="h-10 w-full" />
          <Skeleton className="mt-4 h-10 w-full" />
          <Skeleton className="mt-4 h-10 w-2/3" />
        </div>
      </div>
    )
  }

  if (editing && loadError) {
    return (
      <div className="space-y-6">
        <ErrorState message={loadError} onRetry={reload} />
        <Link to="/dashboard" className="inline-flex items-center gap-1 text-sm font-semibold text-brand hover:underline">
          <Icon name="chevron_left" size={16} />
          Volver a mis solicitudes
        </Link>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <header>
        <Link
          to="/dashboard"
          className="inline-flex items-center gap-1 text-sm font-semibold text-brand hover:underline"
        >
          <Icon name="chevron_left" size={16} />
          Mis solicitudes
        </Link>
        <h1 className="mt-2 text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">
          {editing ? 'Editar solicitud' : 'Nueva solicitud de insumo'}
        </h1>
        <p className="mt-1 text-sm text-gray-500">
          Describe qué necesitas, cuánto y para cuándo. Los agricultores podrán ofertarte.
        </p>
      </header>

      <form
        onSubmit={handleSubmit}
        className="rounded-2xl border border-gray-100 bg-white p-6"
      >
        {error && (
          <p role="alert" className="mb-4 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">
            {error}
          </p>
        )}

        <div className="grid gap-5 md:grid-cols-2">
          <div className="md:col-span-2">
            <label htmlFor="product_name" className={LABEL}>
              Producto o insumo
            </label>
            <input
              id="product_name"
              type="text"
              required
              value={form.product_name}
              onChange={(e) => setField('product_name', e.target.value)}
              placeholder="Ej. Fertilizante urea"
              className={INPUT}
            />
          </div>

          <div>
            <label htmlFor="total_amount" className={LABEL}>
              Monto total (C$)
            </label>
            <input
              id="total_amount"
              type="number"
              min="1"
              step="any"
              required
              value={form.total_amount}
              onChange={(e) => setField('total_amount', e.target.value)}
              className={INPUT}
            />
          </div>

          <div>
            <label htmlFor="unit_of_measure" className={LABEL}>
              Unidad de medida
            </label>
            <select
              id="unit_of_measure"
              value={form.unit_of_measure}
              onChange={(e) => setField('unit_of_measure', e.target.value)}
              className={INPUT}
            >
              {MEASUREMENT_LABELS.map((label, index) => (
                <option key={label} value={String(index)}>
                  {label}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label htmlFor="Department" className={LABEL}>
              Departamento
            </label>
            <input
              id="Department"
              type="text"
              required
              value={form.Department}
              onChange={(e) => setField('Department', e.target.value)}
              placeholder="Ej. Managua"
              className={INPUT}
            />
          </div>

          <div>
            <label htmlFor="Municipality" className={LABEL}>
              Municipio
            </label>
            <input
              id="Municipality"
              type="text"
              value={form.Municipality}
              onChange={(e) => setField('Municipality', e.target.value)}
              placeholder="Ej. Tipitapa"
              className={INPUT}
            />
          </div>

          <div className="md:col-span-2">
            <label htmlFor="AddressLine" className={LABEL}>
              Dirección de entrega
            </label>
            <input
              id="AddressLine"
              type="text"
              value={form.AddressLine}
              onChange={(e) => setField('AddressLine', e.target.value)}
              placeholder="Ej. Carretera a Masaya, km 12"
              className={INPUT}
            />
          </div>

          <div>
            <label htmlFor="request_deadline" className={LABEL}>
              Fecha límite de solicitud
            </label>
            <input
              id="request_deadline"
              type="datetime-local"
              value={form.request_deadline}
              onChange={(e) => setField('request_deadline', e.target.value)}
              className={INPUT}
            />
          </div>

          <div>
            <label htmlFor="delivery_deadline" className={LABEL}>
              Fecha límite de entrega
            </label>
            <input
              id="delivery_deadline"
              type="datetime-local"
              value={form.delivery_deadline}
              onChange={(e) => setField('delivery_deadline', e.target.value)}
              className={INPUT}
            />
          </div>

          <div className="md:col-span-2">
            <label htmlFor="description" className={LABEL}>
              Detalles
            </label>
            <textarea
              id="description"
              rows={3}
              value={form.description}
              onChange={(e) => setField('description', e.target.value)}
              placeholder="Presentación, calidad, condiciones de entrega…"
              className={INPUT}
            />
          </div>

          <div>
            <label htmlFor="min_amount_per_provider" className={LABEL}>
              Monto mínimo por proveedor (C$)
            </label>
            <input
              id="min_amount_per_provider"
              type="number"
              min="0"
              step="any"
              value={form.min_amount_per_provider}
              onChange={(e) => setField('min_amount_per_provider', e.target.value)}
              className={INPUT}
            />
          </div>

          <label className="mt-6 flex items-center gap-3 text-sm text-gray-700">
            <input
              type="checkbox"
              checked={form.multiple_providers}
              onChange={(e) => setField('multiple_providers', e.target.checked)}
              className="h-4 w-4 rounded border-gray-300 text-brand focus:ring-brand"
            />
            Permitir varios proveedores
          </label>
        </div>

        <div className="mt-6 flex flex-wrap justify-end gap-3">
          <Button
            variant="outline"
            onClick={() => navigate(editing ? `/dashboard/requests/${id}` : '/dashboard')}
            disabled={saving}
          >
            Cancelar
          </Button>
          <Button type="submit" disabled={saving}>
            {saving ? 'Guardando…' : editing ? 'Guardar cambios' : 'Crear solicitud'}
          </Button>
        </div>
      </form>
    </div>
  )
}
