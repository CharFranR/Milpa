import { useState } from 'react'
import Icon from '../../components/ui/Icon'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'
import { producerProducts } from '../../mocks/producer'

export default function ProducerProducts() {
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({
    name: '',
    price: '',
    unit: 'kg',
    quantity: '',
  })

  function setField(key, value) {
    setForm((f) => ({ ...f, [key]: value }))
  }

  function handleCancel() {
    setForm({ name: '', price: '', unit: 'kg', quantity: '' })
    setShowForm(false)
  }

  function handleSubmit(e) {
    e.preventDefault()
    setForm({ name: '', price: '', unit: 'kg', quantity: '' })
    setShowForm(false)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900">Mis productos</h1>
        <Button
          type="button"
          variant="primary"
          size="sm"
          onClick={() => setShowForm(!showForm)}
          icon={<Icon name="add" size={16} />}
        >
          Agregar producto
        </Button>
      </div>

      {showForm && (
        <form onSubmit={handleSubmit} className="rounded-xl border border-gray-200 bg-brand-soft/50 p-6 space-y-4">
          <h2 className="text-lg font-semibold text-gray-900">Nuevo producto</h2>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="sm:col-span-2">
              <label htmlFor="prod-name" className="text-xs font-semibold text-gray-600">
                Nombre del producto
              </label>
              <input
                id="prod-name"
                type="text"
                required
                value={form.name}
                onChange={(e) => setField('name', e.target.value)}
                placeholder="Tomates Cherry Orgánicos"
                className="mt-1.5 w-full rounded-lg border border-gray-300 bg-white px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
              />
            </div>
            <div>
              <label htmlFor="prod-price" className="text-xs font-semibold text-gray-600">
                Precio por unidad
              </label>
              <input
                id="prod-price"
                type="number"
                required
                value={form.price}
                onChange={(e) => setField('price', e.target.value)}
                placeholder="12000"
                className="mt-1.5 w-full rounded-lg border border-gray-300 bg-white px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
              />
            </div>
            <div>
              <label htmlFor="prod-unit" className="text-xs font-semibold text-gray-600">
                Unidad de medida
              </label>
              <select
                id="prod-unit"
                value={form.unit}
                onChange={(e) => setField('unit', e.target.value)}
                className="mt-1.5 w-full rounded-lg border border-gray-300 bg-white px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
              >
                <option value="kg">Kilogramo (kg)</option>
                <option value="un">Unidad (un)</option>
                <option value="bandeja">Bandeja</option>
                <option value="docena">Docena</option>
                <option value="500 ml">500 ml</option>
                <option value="250 g">250 g</option>
                <option value="atado">Atado</option>
              </select>
            </div>
            <div>
              <label htmlFor="prod-qty" className="text-xs font-semibold text-gray-600">
                Cantidad disponible
              </label>
              <input
                id="prod-qty"
                type="number"
                required
                value={form.quantity}
                onChange={(e) => setField('quantity', e.target.value)}
                placeholder="100"
                className="mt-1.5 w-full rounded-lg border border-gray-300 bg-white px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
              />
            </div>
          </div>
          <div className="flex justify-end gap-3 pt-2">
            <Button type="button" variant="outline" onClick={handleCancel}>
              Cancelar
            </Button>
            <Button type="submit" variant="primary">
              Publicar producto
            </Button>
          </div>
        </form>
      )}

      <div className="grid gap-5 sm:grid-cols-2 xl:grid-cols-3">
        {producerProducts.map((prod) => (
          <article
            key={prod.id}
            className="group relative flex flex-col overflow-hidden rounded-2xl border border-gray-100 bg-white transition-shadow hover:shadow-lg"
          >
            <div className="relative aspect-[4/3] bg-gray-100">
              <span
                className="absolute left-3 top-3"
                role="img"
                aria-label="Imagen del producto"
              >
                <span className="flex h-32 w-32 items-center justify-center rounded-xl bg-gray-200 text-gray-400 text-4xl">📦</span>
              </span>
              <Badge
                className="absolute right-3 top-3"
                tone={prod.available ? 'brand' : 'amber'}
              >
                {prod.available ? 'Activo' : 'Agotado'}
              </Badge>
            </div>
            <div className="flex flex-1 flex-col p-4">
              <h3 className="font-semibold text-gray-900 line-clamp-1">{prod.name}</h3>
              <p className="mt-2 text-lg font-bold text-brand">
                ${prod.price.toLocaleString()}
                <span className="text-sm font-medium text-gray-400"> / {prod.unit}</span>
              </p>
              <div className="mt-4 flex items-center gap-2">
                <Button variant="outline" size="sm" className="flex-1">
                  Ver
                </Button>
                <Button variant="primary" size="sm" className="flex-1">
                  Editar
                </Button>
              </div>
            </div>
          </article>
        ))}
      </div>
    </div>
  )
}