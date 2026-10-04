import { useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import Button from '../../components/ui/Button'
import ErrorState from '../../components/ui/ErrorState'
import Icon from '../../components/ui/Icon'
import Skeleton from '../../components/ui/Skeleton'
import { useSupplyRequest } from '../../hooks/useSupplyRequests'
import { formatPrice } from '../../lib/format'
import { formatDeadline, measurementLabel } from '../../lib/supplyStatus'
import { supplyOffers } from '../../services/supplyOffers'

const INPUT =
  'mt-1.5 w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand'
const LABEL = 'text-xs font-semibold text-gray-600'

const EMPTY_FORM = {
  total_amount: '',
  price_per_unit: '',
  measurement: '0',
  delivery_day: '',
  delivery_available: true,
  comments: '',
}

export default function ProducerOfferForm() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { request, loading, error: loadError, reload } = useSupplyRequest(id)

  const [form, setForm] = useState(EMPTY_FORM)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  function setField(field, value) {
    setForm((prev) => ({ ...prev, [field]: value }))
  }

  async function handleSubmit(event) {
    event.preventDefault()

    const total = Number(form.total_amount)
    const price = Number(form.price_per_unit)
    if (!(total > 0)) {
      setError('El monto de tu oferta debe ser mayor que cero.')
      return
    }
    if (!(price > 0)) {
      setError('El precio por unidad debe ser mayor que cero.')
      return
    }
    if (request && total > request.actual_amount) {
      setError(
        `Tu oferta supera lo disponible (${formatPrice(request.actual_amount)}).`,
      )
      return
    }

    setSaving(true)
    setError('')
    try {
      await supplyOffers.create({
        supply_request_id: id,
        total_amount: total,
        price_per_unit: price,
        measurement: Number(form.measurement),
        delivery_day: form.delivery_day
          ? new Date(form.delivery_day).toISOString()
          : undefined,
        delivery_available: form.delivery_available,
        comments: form.comments,
      })
      navigate('/producer/offers')
    } catch (err) {
      const duplicated = err.status === 409 && /already exists/i.test(err.message || '')
      setError(
        duplicated
          ? 'Ya enviaste una oferta para esta solicitud: no se puede repetir, ni siquiera después de retirarla.'
          : err.message || 'No se pudo enviar la oferta.',
      )
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-6 w-44" />
        <div className="rounded-2xl border border-gray-100 bg-white p-6">
          <Skeleton className="h-5 w-64" />
          <Skeleton className="mt-3 h-4 w-80" />
        </div>
        <div className="rounded-2xl border border-gray-100 bg-white p-6">
          <Skeleton className="h-10 w-full" />
          <Skeleton className="mt-4 h-10 w-2/3" />
        </div>
      </div>
    )
  }

  if (loadError) {
    return (
      <div className="space-y-6">
        <ErrorState message={loadError} onRetry={reload} />
        <Link
          to="/producer/available"
          className="inline-flex items-center gap-1 text-sm font-semibold text-brand hover:underline"
        >
          <Icon name="chevron_left" size={16} />
          Solicitudes disponibles
        </Link>
      </div>
    )
  }

  if (!request) return null

  return (
    <div className="space-y-6">
      <header>
        <Link
          to="/producer/available"
          className="inline-flex items-center gap-1 text-sm font-semibold text-brand hover:underline"
        >
          <Icon name="chevron_left" size={16} />
          Solicitudes disponibles
        </Link>
        <h1 className="mt-2 text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">
          Ofrecer en «{request.product_name}»
        </h1>
        <p className="mt-1 text-sm text-gray-500">
          El comprador verá tu oferta junto a las de otros proveedores.
        </p>
      </header>

      <section className="rounded-2xl border border-gray-100 bg-white p-6">
        <h2 className="text-base font-bold text-gray-900">Lo que piden</h2>
        <div className="mt-3 flex flex-wrap gap-x-6 gap-y-2 text-sm text-gray-600">
          <span>
            <strong className="text-gray-900">{formatPrice(request.total_amount)}</strong> en total
          </span>
          <span>
            <strong className="text-gray-900">{formatPrice(request.actual_amount)}</strong>{' '}
            disponibles
          </span>
          <span>Unidad: {measurementLabel(request.unit_of_measure)}</span>
          <span>Entrega hasta: {formatDeadline(request.delivery_deadline)}</span>
          <span>{request.address?.Department || 'Sin departamento'}</span>
        </div>
        {request.status !== 0 && (
          <p className="mt-3 rounded-lg bg-amber-50 px-3 py-2 text-sm text-amber-800">
            Esta solicitud ya no está abierta: no se aceptarán más ofertas.
          </p>
        )}
      </section>

      <form onSubmit={handleSubmit} className="rounded-2xl border border-gray-100 bg-white p-6">
        {error && (
          <p role="alert" className="mb-4 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">
            {error}
          </p>
        )}

        <div className="grid gap-5 md:grid-cols-2">
          <div>
            <label htmlFor="total_amount" className={LABEL}>
              Monto total de tu oferta (C$)
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
            <label htmlFor="price_per_unit" className={LABEL}>
              Precio por unidad (C$)
            </label>
            <input
              id="price_per_unit"
              type="number"
              min="1"
              step="any"
              required
              value={form.price_per_unit}
              onChange={(e) => setField('price_per_unit', e.target.value)}
              className={INPUT}
            />
          </div>

          <div>
            <label htmlFor="measurement" className={LABEL}>
              Unidad de medida
            </label>
            <select
              id="measurement"
              value={form.measurement}
              onChange={(e) => setField('measurement', e.target.value)}
              className={INPUT}
            >
              <option value="0">Kg</option>
              <option value="1">Lb</option>
              <option value="2">Tn</option>
            </select>
          </div>

          <div>
            <label htmlFor="delivery_day" className={LABEL}>
              Día estimado de entrega
            </label>
            <input
              id="delivery_day"
              type="date"
              value={form.delivery_day}
              onChange={(e) => setField('delivery_day', e.target.value)}
              className={INPUT}
            />
          </div>

          <div className="md:col-span-2">
            <label htmlFor="comments" className={LABEL}>
              Comentarios
            </label>
            <textarea
              id="comments"
              rows={3}
              value={form.comments}
              onChange={(e) => setField('comments', e.target.value)}
              placeholder="Condiciones de pago, calidad, logística…"
              className={INPUT}
            />
          </div>

          <label className="flex items-center gap-3 text-sm text-gray-700">
            <input
              type="checkbox"
              checked={form.delivery_available}
              onChange={(e) => setField('delivery_available', e.target.checked)}
              className="h-4 w-4 rounded border-gray-300 text-brand focus:ring-brand"
            />
            Puedo entregar en la fecha solicitada
          </label>
        </div>

        <div className="mt-6 flex flex-wrap justify-end gap-3">
          <Button variant="outline" onClick={() => navigate('/producer/available')} disabled={saving}>
            Cancelar
          </Button>
          <Button type="submit" disabled={saving || request.status !== 0}>
            {saving ? 'Enviando…' : 'Enviar oferta'}
          </Button>
        </div>
      </form>
    </div>
  )
}
