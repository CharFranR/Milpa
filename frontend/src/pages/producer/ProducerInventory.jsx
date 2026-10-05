import { useState } from 'react'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'
import ConfirmDialog from '../../components/ui/ConfirmDialog'
import EmptyState from '../../components/ui/EmptyState'
import ErrorState from '../../components/ui/ErrorState'
import Icon from '../../components/ui/Icon'
import Skeleton from '../../components/ui/Skeleton'
import Toast from '../../components/ui/Toast'
import { useAuth } from '../../context/AuthContext'
import { useAvailability, useInventory } from '../../hooks/useInventory'
import { formatDateTime, measurementLabel } from '../../lib/supplyStatus'
import { inventory } from '../../services/inventory'

const INPUT =
  'mt-1.5 w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand'
const LABEL = 'text-xs font-semibold text-gray-600'

const EMPTY_FORM = { product_name: '', quantity: '', measurement: '0' }

export default function ProducerInventory() {
  const { user } = useAuth()
  const { items, loading, error, reload } = useInventory()
  const availability = useAvailability(user?.id, items)

  const [form, setForm] = useState(EMPTY_FORM)
  const [open, setOpen] = useState(false)
  const [editingId, setEditingId] = useState(null)
  const [saving, setSaving] = useState(false)
  const [formError, setFormError] = useState('')
  const [dialog, setDialog] = useState(null)
  const [toast, setToast] = useState(null)

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
      measurement: String(item.measurement ?? 0),
    })
    setFormError('')
    setOpen(true)
  }

  async function handleSubmit(event) {
    event.preventDefault()
    if (!form.product_name.trim()) {
      setFormError('Indica el nombre del producto.')
      return
    }
    if (Number(form.quantity) < 0) {
      setFormError('La existencia no puede ser negativa.')
      return
    }

    setSaving(true)
    setFormError('')
    try {
      await inventory.upsert({
        product_name: form.product_name.trim(),
        quantity: Number(form.quantity),
        measurement: Number(form.measurement),
      })
      setOpen(false)
      setEditingId(null)
      setForm(EMPTY_FORM)
      setToast({ message: 'Inventario actualizado.', tone: 'success' })
      reload()
    } catch (err) {
      setFormError(err.message || 'No se pudo guardar el producto.')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(item) {
    setSaving(true)
    try {
      await inventory.remove(item.id)
      setDialog(null)
      setToast({ message: 'Producto eliminado del inventario.', tone: 'success' })
      reload()
    } catch (err) {
      setDialog(null)
      setToast({ message: err.message || 'No se pudo eliminar el producto.', tone: 'error' })
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="space-y-6">
      <header className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">
            Mi inventario
          </h1>
          <p className="mt-1 text-sm text-gray-500">
            Existencias que usas para calcular lo que puedes ofrecer en cada solicitud.
          </p>
        </div>
        <Button
          onClick={open && !editingId ? () => setOpen(false) : startCreate}
          icon={<Icon name={open && !editingId ? 'close' : 'add'} size={18} />}
        >
          {open && !editingId ? 'Cerrar' : 'Agregar producto'}
        </Button>
      </header>

      {open && (
        <form onSubmit={handleSubmit} className="rounded-2xl border border-gray-100 bg-white p-6">
          <h2 className="text-base font-bold text-gray-900">
            {editingId ? 'Editar producto' : 'Agregar producto'}
          </h2>

          {formError && (
            <p role="alert" className="mt-4 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">
              {formError}
            </p>
          )}

          <div className="mt-4 grid gap-5 md:grid-cols-3">
            <div>
              <label htmlFor="product_name" className={LABEL}>
                Producto
              </label>
              <input
                id="product_name"
                type="text"
                required
                readOnly={Boolean(editingId)}
                value={form.product_name}
                onChange={(e) => setField('product_name', e.target.value)}
                placeholder="Ej. Fertilizante urea"
                className={INPUT}
              />
              {editingId && (
                <p className="mt-1 text-xs text-gray-400">
                  El nombre identifica la fila: no se puede cambiar.
                </p>
              )}
            </div>

            <div>
              <label htmlFor="quantity" className={LABEL}>
                Existencias
              </label>
              <input
                id="quantity"
                type="number"
                min="0"
                step="any"
                required
                value={form.quantity}
                onChange={(e) => setField('quantity', e.target.value)}
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
              <Skeleton className="h-5 w-52" />
            </div>
          ))}
        </div>
      )}

      {!loading && error && <ErrorState message={error} onRetry={reload} />}

      {!loading && !error && items.length === 0 && (
        <EmptyState
          icon="warehouse"
          title="Inventario vacío"
          description="Agrega tus productos y existencias para que el sistema calcule tu disponibilidad."
          action={<Button onClick={startCreate}>Agregar producto</Button>}
        />
      )}

      {!loading && !error && items.length > 0 && (
        <div className="overflow-x-auto rounded-2xl border border-gray-100 bg-white">
          <table className="w-full min-w-[640px] text-left text-sm">
            <thead>
              <tr className="border-b border-gray-100 text-xs uppercase tracking-wide text-gray-400">
                <th className="px-5 py-3 font-semibold">Producto</th>
                <th className="px-5 py-3 font-semibold">Existencias</th>
                <th className="px-5 py-3 font-semibold">Disponible</th>
                <th className="px-5 py-3 font-semibold">Unidad</th>
                <th className="px-5 py-3 font-semibold">Actualizado</th>
                <th className="px-5 py-3 text-right font-semibold">Acciones</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <tr key={item.id} className="border-b border-gray-50 last:border-0">
                  <td className="px-5 py-3 font-medium text-gray-900">{item.product_name}</td>
                  <td className="px-5 py-3 text-gray-700">{item.quantity}</td>
                  <td className="px-5 py-3 text-gray-700">
                    {availability[item.product_name] === undefined ? (
                      <span className="text-gray-300">…</span>
                    ) : availability[item.product_name] === null ? (
                      <span className="text-gray-400">—</span>
                    ) : (
                      availability[item.product_name]
                    )}
                  </td>
                  <td className="px-5 py-3">
                    <Badge tone="brand">{measurementLabel(item.measurement)}</Badge>
                  </td>
                  <td className="px-5 py-3 text-gray-500">{formatDateTime(item.updated_at)}</td>
                  <td className="px-5 py-3">
                    <div className="flex justify-end gap-2">
                      <Button variant="ghost" size="sm" onClick={() => startEdit(item)}>
                        Editar
                      </Button>
                      <Button variant="danger" size="sm" onClick={() => setDialog(item)}>
                        Eliminar
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <ConfirmDialog
        open={Boolean(dialog)}
        danger
        title="¿Eliminar del inventario?"
        message={`Se borrará «${dialog?.product_name || ''}» y su disponibilidad dejará de contar para las recomendaciones.`}
        confirmLabel="Sí, eliminar"
        loading={saving}
        onConfirm={() => handleDelete(dialog)}
        onCancel={() => setDialog(null)}
      />

      <Toast message={toast?.message} tone={toast?.tone} onClose={() => setToast(null)} />
    </div>
  )
}
