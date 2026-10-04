import { useCallback, useEffect, useState } from 'react'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'
import ConfirmDialog from '../../components/ui/ConfirmDialog'
import EmptyState from '../../components/ui/EmptyState'
import ErrorState from '../../components/ui/ErrorState'
import Icon from '../../components/ui/Icon'
import Skeleton from '../../components/ui/Skeleton'
import Toast from '../../components/ui/Toast'
import { useAuth } from '../../context/AuthContext'
import { formatDateTime } from '../../lib/supplyStatus'
import { formatPrice } from '../../lib/format'
import { hasExpired, remainingLabel, statusLabel, toLocalInputValue } from '../../lib/liquidations'
import { liquidations } from '../../services/liquidations'

const INPUT =
  'mt-1.5 w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand'
const LABEL = 'text-xs font-semibold text-gray-600'

const EMPTY_FORM = {
  product_name: '',
  quantity: '',
  unit_of_measure: '',
  total_price: '',
  unit_price: '',
  delivery_time: '',
  visibility: 'public',
  expires_at: '',
}

const STATUS_TONE = {
  Abierta: 'green',
  Cerrada: 'gray',
  Expirada: 'red',
  Asignada: 'brand',
}

function statusTone(label) {
  return STATUS_TONE[label] || 'gray'
}

export default function ProducerLiquidations() {
  const { user } = useAuth()
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const [open, setOpen] = useState(false)
  const [editingId, setEditingId] = useState(null)
  const [form, setForm] = useState(EMPTY_FORM)
  const [formError, setFormError] = useState('')
  const [saving, setSaving] = useState(false)
  const [dialog, setDialog] = useState(null)
  const [toast, setToast] = useState(null)

  const load = useCallback(async () => {
    if (!user?.id) return
    setLoading(true)
    setError('')
    try {
      const data = await liquidations.getBySupplier(user.id)
      setItems(Array.isArray(data) ? data : [])
    } catch (err) {
      setError(err.message || 'No se pudieron cargar tus liquidaciones.')
    } finally {
      setLoading(false)
    }
  }, [user?.id])

  useEffect(() => {
    load()
  }, [load])

  function setField(field, value) {
    setForm((prev) => ({ ...prev, [field]: value }))
  }

  function startCreate() {
    setEditingId(null)
    setForm(EMPTY_FORM)
    setFormError('')
    setOpen(true)
  }

  function startEdit(item) {
    setEditingId(item.id)
    setForm({
      product_name: item.product_name,
      quantity: String(item.quantity),
      unit_of_measure: item.unit_of_measure,
      total_price: String(item.total_price),
      unit_price: String(item.unit_price),
      delivery_time: item.delivery_time || '',
      visibility: item.visibility,
      expires_at: toLocalInputValue(item.expires_at),
    })
    setFormError('')
    setOpen(true)
  }

  function validate() {
    if (!form.product_name.trim()) return 'Indica el nombre del producto.'
    if (!(Number(form.quantity) > 0)) return 'La cantidad debe ser mayor que cero.'
    if (!form.unit_of_measure.trim()) return 'Indica la unidad de medida.'
    if (!(Number(form.total_price) > 0) || !(Number(form.unit_price) > 0)) {
      return 'Los precios deben ser mayores que cero.'
    }
    if (form.expires_at && new Date(form.expires_at).getTime() <= Date.now()) {
      return 'La fecha de vencimiento debe ser futura.'
    }
    return ''
  }

  function buildPayload() {
    const payload = {
      product_name: form.product_name.trim(),
      quantity: Number(form.quantity),
      unit_of_measure: form.unit_of_measure.trim(),
      total_price: Number(form.total_price),
      unit_price: Number(form.unit_price),
      visibility: form.visibility,
    }
    if (form.delivery_time.trim()) payload.delivery_time = form.delivery_time.trim()
    if (form.expires_at) payload.expires_at = new Date(form.expires_at).toISOString()
    return payload
  }

  async function handleSubmit(event) {
    event.preventDefault()

    const problem = validate()
    if (problem) {
      setFormError(problem)
      return
    }

    const payload = buildPayload()
    setSaving(true)
    setFormError('')
    try {
      if (editingId) {
        await liquidations.update(editingId, payload)
        setItems((prev) =>
          prev.map((item) => (item.id === editingId ? { ...item, ...payload } : item)),
        )
        setToast({ message: 'Liquidación actualizada.', tone: 'success' })
      } else {
        const created = await liquidations.create(payload)
        setItems((prev) => [created, ...prev.filter((item) => item.id !== created.id)])
        setToast({
          message:
            created.visibility === 'private'
              ? 'Liquidación privada guardada. El listado del servidor devuelve las públicas, así que esta fila se ve solo hasta que recargues.'
              : 'Liquidación publicada.',
          tone: 'success',
        })
      }
      setOpen(false)
      setEditingId(null)
      setForm(EMPTY_FORM)
    } catch (err) {
      setFormError(err.message || 'No se pudo guardar la liquidación.')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(item) {
    setSaving(true)
    try {
      await liquidations.remove(item.id)
      setDialog(null)
      setItems((prev) => prev.filter((entry) => entry.id !== item.id))
      setToast({ message: 'Liquidación eliminada.', tone: 'success' })
    } catch (err) {
      setDialog(null)
      setToast({ message: err.message || 'No se pudo eliminar.', tone: 'error' })
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="space-y-6">
      <header className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">
            Liquidaciones
          </h1>
          <p className="mt-1 text-sm text-gray-500">
            Lotes de excedente de cosecha a precio preferencial para evitar la pérdida total del
            producto.
          </p>
        </div>
        <Button
          onClick={open && !editingId ? () => setOpen(false) : startCreate}
          icon={<Icon name={open && !editingId ? 'close' : 'add'} size={18} />}
        >
          {open && !editingId ? 'Cerrar' : 'Publicar liquidación'}
        </Button>
      </header>

      {open && (
        <form onSubmit={handleSubmit} className="rounded-2xl border border-gray-100 bg-white p-6">
          <h2 className="text-base font-bold text-gray-900">
            {editingId ? 'Editar liquidación' : 'Publicar liquidación'}
          </h2>

          {formError && (
            <p role="alert" className="mt-4 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">
              {formError}
            </p>
          )}

          <div className="mt-4 grid gap-5 md:grid-cols-3">
            <div>
              <label htmlFor="liq_product" className={LABEL}>
                Producto
              </label>
              <input
                id="liq_product"
                type="text"
                required
                value={form.product_name}
                onChange={(e) => setField('product_name', e.target.value)}
                placeholder="Ej. Maíz amarillo"
                className={INPUT}
              />
            </div>

            <div>
              <label htmlFor="liq_quantity" className={LABEL}>
                Cantidad
              </label>
              <input
                id="liq_quantity"
                type="number"
                min="0.01"
                step="any"
                required
                value={form.quantity}
                onChange={(e) => setField('quantity', e.target.value)}
                placeholder="Ej. 500"
                className={INPUT}
              />
            </div>

            <div>
              <label htmlFor="liq_unit" className={LABEL}>
                Unidad de medida
              </label>
              <input
                id="liq_unit"
                type="text"
                required
                value={form.unit_of_measure}
                onChange={(e) => setField('unit_of_measure', e.target.value)}
                placeholder="Ej. kg, quintales, t"
                className={INPUT}
              />
            </div>

            <div>
              <label htmlFor="liq_total" className={LABEL}>
                Precio total del lote (C$)
              </label>
              <input
                id="liq_total"
                type="number"
                min="0.01"
                step="any"
                required
                value={form.total_price}
                onChange={(e) => setField('total_price', e.target.value)}
                placeholder="Ej. 12500"
                className={INPUT}
              />
            </div>

            <div>
              <label htmlFor="liq_unit_price" className={LABEL}>
                Precio por unidad (C$)
              </label>
              <input
                id="liq_unit_price"
                type="number"
                min="0.01"
                step="any"
                required
                value={form.unit_price}
                onChange={(e) => setField('unit_price', e.target.value)}
                placeholder="Ej. 25"
                className={INPUT}
              />
            </div>

            <div>
              <label htmlFor="liq_delivery" className={LABEL}>
                Tiempo de entrega
              </label>
              <input
                id="liq_delivery"
                type="text"
                value={form.delivery_time}
                onChange={(e) => setField('delivery_time', e.target.value)}
                placeholder="Ej. 3 días"
                className={INPUT}
              />
            </div>

            <div>
              <label htmlFor="liq_visibility" className={LABEL}>
                Visibilidad
              </label>
              <select
                id="liq_visibility"
                value={form.visibility}
                onChange={(e) => setField('visibility', e.target.value)}
                className={INPUT}
              >
                <option value="public">Todos los compradores</option>
                <option value="private">Solo mayoristas (privada)</option>
              </select>
            </div>

            <div>
              <label htmlFor="liq_expires" className={LABEL}>
                Vence (opcional)
              </label>
              <input
                id="liq_expires"
                type="datetime-local"
                value={form.expires_at}
                onChange={(e) => setField('expires_at', e.target.value)}
                className={INPUT}
              />
              <p className="mt-1 text-xs text-gray-400">
                Pasada esa fecha deja de mostrarse en el listado público.
              </p>
            </div>
          </div>

          <div className="mt-6 flex justify-end gap-3">
            <Button
              variant="outline"
              onClick={() => {
                setOpen(false)
                setEditingId(null)
                setFormError('')
              }}
              disabled={saving}
            >
              Cancelar
            </Button>
            <Button type="submit" disabled={saving}>
              {saving ? 'Guardando…' : 'Guardar'}
            </Button>
          </div>
        </form>
      )}

      {loading && (
        <div className="space-y-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="rounded-xl border border-gray-100 bg-white p-5">
              <Skeleton className="h-5 w-64" />
            </div>
          ))}
        </div>
      )}

      {!loading && error && <ErrorState message={error} onRetry={load} />}

      {!loading && !error && items.length === 0 && (
        <EmptyState
          icon="sell"
          title="Sin liquidaciones"
          description="Publica un lote de excedente a precio preferencial para evitar que la cosecha se pierda."
          action={<Button onClick={startCreate}>Publicar liquidación</Button>}
        />
      )}

      {!loading && !error && items.length > 0 && (
        <div className="overflow-x-auto rounded-2xl border border-gray-100 bg-white">
          <table className="w-full min-w-[860px] text-left text-sm">
            <thead>
              <tr className="border-b border-gray-100 text-xs uppercase tracking-wide text-gray-400">
                <th className="px-5 py-3 font-semibold">Producto</th>
                <th className="px-5 py-3 font-semibold">Cantidad</th>
                <th className="px-5 py-3 font-semibold">Total</th>
                <th className="px-5 py-3 font-semibold">Unitario</th>
                <th className="px-5 py-3 font-semibold">Visibilidad</th>
                <th className="px-5 py-3 font-semibold">Vencimiento</th>
                <th className="px-5 py-3 font-semibold">Estado</th>
                <th className="px-5 py-3 text-right font-semibold">Acciones</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {items.map((item) => {
                const label = statusLabel(item)
                const expiredByTime = hasExpired(item)
                return (
                  <tr key={item.id} className="hover:bg-gray-50/60">
                    <td className="px-5 py-3.5 font-semibold text-gray-900">{item.product_name}</td>
                    <td className="px-5 py-3.5 text-gray-600">
                      {item.quantity} {item.unit_of_measure}
                    </td>
                    <td className="px-5 py-3.5 font-semibold text-brand">
                      {formatPrice(item.total_price)}
                    </td>
                    <td className="px-5 py-3.5 text-gray-600">{formatPrice(item.unit_price)}</td>
                    <td className="px-5 py-3.5">
                      <Badge tone={item.visibility === 'private' ? 'amber' : 'brand'}>
                        {item.visibility === 'private' ? 'Privada' : 'Pública'}
                      </Badge>
                    </td>
                    <td className="px-5 py-3.5 text-gray-600">
                      {item.expires_at ? (
                        <>
                          <span className="block">{formatDateTime(item.expires_at)}</span>
                          <span
                            className={`text-xs ${expiredByTime ? 'text-red-600' : 'text-gray-400'}`}
                          >
                            {remainingLabel(item.expires_at)}
                          </span>
                        </>
                      ) : (
                        <span className="text-gray-400">Sin fecha</span>
                      )}
                    </td>
                    <td className="px-5 py-3.5">
                      <Badge tone={statusTone(label)}>{label}</Badge>
                    </td>
                    <td className="px-5 py-3.5">
                      <div className="flex justify-end gap-2">
                        <button
                          type="button"
                          onClick={() => startEdit(item)}
                          aria-label={`Editar ${item.product_name}`}
                          className="inline-flex h-8 w-8 items-center justify-center rounded-lg text-gray-500 transition-colors hover:bg-brand-soft hover:text-brand"
                        >
                          <Icon name="edit" size={17} />
                        </button>
                        <button
                          type="button"
                          onClick={() => setDialog(item)}
                          aria-label={`Eliminar ${item.product_name}`}
                          className="inline-flex h-8 w-8 items-center justify-center rounded-lg text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600"
                        >
                          <Icon name="delete" size={17} />
                        </button>
                      </div>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}

      <ConfirmDialog
        open={Boolean(dialog)}
        danger
        title="Eliminar liquidación"
        message={`Se eliminará «${dialog?.product_name || ''}» de forma permanente.`}
        confirmLabel="Sí, eliminar"
        loading={saving}
        onConfirm={() => handleDelete(dialog)}
        onCancel={() => setDialog(null)}
      />

      {toast && <Toast message={toast.message} tone={toast.tone} onClose={() => setToast(null)} />}
    </div>
  )
}
