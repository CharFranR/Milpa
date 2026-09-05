import { useState, useEffect, useRef } from 'react'
import Icon from '../../components/ui/Icon'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'
import { useOfferings } from '../../hooks/useOfferings'
import { categories } from '../../services/api'
import { getCompanyId } from '../../lib/session'
import { formatPrice } from '../../lib/format'
import { setProductImage, getProductImage } from '../../lib/productImages'

const MAX_IMAGE_SIZE = 5 * 1024 * 1024

export default function ProducerProducts() {
  const companyId = getCompanyId()
  const { offeringsList, loading, error, createOffering, updateOffering } = useOfferings(companyId)
  const [showForm, setShowForm] = useState(false)
  const [editingId, setEditingId] = useState(null)
  const [saving, setSaving] = useState(false)
  const [formError, setFormError] = useState('')
  const [cats, setCats] = useState([])
  const [form, setForm] = useState({
    name: '',
    price: '',
    unit: 'kg',
    quantity: '',
    category: '',
    description: '',
    image_url: '',
  })
  const fileRef = useRef(null)
  const [imagePreview, setImagePreview] = useState('')

  useEffect(() => {
    categories.getAll().then((data) => setCats(Array.isArray(data) ? data : [])).catch(() => {})
  }, [])

  function setField(key, value) {
    setForm((f) => ({ ...f, [key]: value }))
  }

  function handleImageSelect(e) {
    const file = e.target.files?.[0]
    if (!file) return

    if (file.size > MAX_IMAGE_SIZE) {
      setFormError('La imagen no puede superar 5 MB.')
      return
    }

    const reader = new FileReader()
    reader.onload = () => {
      const dataUrl = reader.result
      setForm((f) => ({ ...f, image_url: dataUrl }))
      setImagePreview(dataUrl)
    }
    reader.readAsDataURL(file)
  }

  function handleRemoveImage() {
    setForm((f) => ({ ...f, image_url: '' }))
    setImagePreview('')
    if (fileRef.current) fileRef.current.value = ''
  }

  function openCreate() {
    setForm({ name: '', price: '', unit: 'kg', quantity: '', category: '', description: '', image_url: '' })
    setImagePreview('')
    setEditingId(null)
    setShowForm(true)
    setFormError('')
  }

  function openEdit(offering) {
    const desc = offering.description || ''
    const unitMatch = desc.match(/Unit:\s*(\S+)/)
    const qtyMatch = desc.match(/Qty:\s*(\d+)/)
    const catMatch = desc.match(/Category:\s*(.+)/)
    const cleanDesc = desc.replace(/Unit:\s*\S+\n?/, '').replace(/Qty:\s*\d+\n?/, '').replace(/Category:\s*.+\n?/, '').trim()

    const savedImage = offering.image_url || getProductImage(offering.id) || ''

    setForm({
      name: offering.name || '',
      price: String(offering.price || ''),
      unit: unitMatch?.[1] || 'kg',
      quantity: qtyMatch?.[1] || '',
      category: catMatch?.[1] || '',
      description: cleanDesc,
      image_url: savedImage,
    })
    setImagePreview(savedImage)
    setEditingId(offering.id)
    setShowForm(true)
    setFormError('')
  }

  function handleCancel() {
    setForm({ name: '', price: '', unit: 'kg', quantity: '', category: '', description: '', image_url: '' })
    setImagePreview('')
    setEditingId(null)
    setShowForm(false)
    setFormError('')
  }

  function handleSubmit(e) {
    e.preventDefault()
    if (!form.name.trim() || !form.price) {
      setFormError('Nombre y precio son obligatorios.')
      return
    }
    if (!companyId) {
      setFormError('Primero debes crear tu empresa en "Mi negocio".')
      return
    }

    setSaving(true)
    setFormError('')

    let descParts = []
    if (form.unit) descParts.push(`Unit: ${form.unit}`)
    if (form.quantity) descParts.push(`Qty: ${form.quantity}`)
    if (form.category) descParts.push(`Category: ${form.category}`)
    if (form.description) descParts.push('')
    if (form.description) descParts.push(form.description)
    const description = descParts.join('\n')

    const payload = {
      company_id: companyId,
      type: 0,
      name: form.name.trim(),
      description,
      price: parseFloat(form.price),
    }

    const action = editingId
      ? updateOffering(editingId, payload)
      : createOffering(payload)

    action
      .then((result) => {
        if (form.image_url && result?.id) {
          setProductImage(result.id, form.image_url)
        }
        handleCancel()
      })
      .catch((err) => {
        setFormError(err.message || 'Error al guardar.')
      })
      .finally(() => setSaving(false))
  }

  if (!companyId) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold text-gray-900">Mis productos</h1>
        </div>
        <div className="rounded-xl border border-dashed border-gray-300 bg-white p-12 text-center">
          <span className="flex h-16 w-16 mx-auto items-center justify-center rounded-2xl bg-brand-soft text-brand">
            <Icon name="inventory_2" size={32} />
          </span>
          <h2 className="mt-4 text-lg font-semibold text-gray-900">Primero crea tu empresa</h2>
          <p className="mt-2 max-w-sm mx-auto text-sm text-gray-500">
            Para publicar productos, necesitas tener una empresa creada.
          </p>
          <Button type="button" variant="primary" className="mt-6" onClick={() => window.location.hash = '#/producer'} icon={<Icon name="storefront" size={18} />}>
            Ir a Mi negocio
          </Button>
        </div>
      </div>
    )
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold text-gray-900">Mis productos</h1>
        </div>
        <div className="grid gap-5 sm:grid-cols-2 xl:grid-cols-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="animate-pulse rounded-2xl border border-gray-100 bg-white p-4 space-y-3">
              <div className="aspect-[4/3] rounded-xl bg-gray-200" />
              <div className="h-4 w-3/4 rounded bg-gray-200" />
              <div className="h-5 w-1/2 rounded bg-gray-200" />
            </div>
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900">Mis productos</h1>
        <Button type="button" variant="primary" size="sm" onClick={openCreate} icon={<Icon name="add" size={16} />}>
          Agregar producto
        </Button>
      </div>

      {error && (
        <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{error}</p>
      )}

      {showForm && (
        <form onSubmit={handleSubmit} className="rounded-xl border border-gray-200 bg-brand-soft/50 p-6 space-y-4">
          <h2 className="text-lg font-semibold text-gray-900">
            {editingId ? 'Editar producto' : 'Nuevo producto'}
          </h2>

          {formError && (
            <p className="rounded-lg bg-red-50 px-3 py-2 text-xs text-red-700">{formError}</p>
          )}

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="sm:col-span-2">
              <label className="text-xs font-semibold text-gray-600">Nombre del producto *</label>
              <input
                type="text"
                required
                value={form.name}
                onChange={(e) => setField('name', e.target.value)}
                placeholder="Tomates Cherry Orgánicos"
                className="mt-1.5 w-full rounded-lg border border-gray-300 bg-white px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
              />
            </div>

            <div className="sm:col-span-2">
              <label className="text-xs font-semibold text-gray-600">Imagen del producto</label>
              <div className="mt-1.5">
                {imagePreview ? (
                  <div className="relative inline-block">
                    <img src={imagePreview} alt="Vista previa" className="h-32 w-32 rounded-xl object-cover border border-gray-200" />
                    <button
                      type="button"
                      onClick={handleRemoveImage}
                      className="absolute -top-2 -right-2 rounded-full bg-red-500 p-1 text-white hover:bg-red-600"
                      aria-label="Eliminar imagen"
                    >
                      <Icon name="close" size={14} />
                    </button>
                  </div>
                ) : (
                  <label className="flex flex-col items-center gap-2 rounded-xl border-2 border-dashed border-gray-300 bg-white p-6 cursor-pointer hover:border-brand/40 hover:bg-brand-soft/20 transition-colors">
                    <Icon name="add_a_photo" size={24} className="text-gray-400" />
                    <span className="text-sm text-gray-500">Subir imagen (máx. 5 MB)</span>
                    <input
                      ref={fileRef}
                      type="file"
                      accept="image/*"
                      onChange={handleImageSelect}
                      className="sr-only"
                    />
                  </label>
                )}
              </div>
            </div>

            <div>
              <label className="text-xs font-semibold text-gray-600">Precio (C$) *</label>
              <input
                type="number"
                required
                value={form.price}
                onChange={(e) => setField('price', e.target.value)}
                placeholder="180"
                className="mt-1.5 w-full rounded-lg border border-gray-300 bg-white px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
              />
            </div>
            <div>
              <label className="text-xs font-semibold text-gray-600">Unidad de medida</label>
              <select
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
              <label className="text-xs font-semibold text-gray-600">Cantidad disponible</label>
              <input
                type="number"
                value={form.quantity}
                onChange={(e) => setField('quantity', e.target.value)}
                placeholder="100"
                className="mt-1.5 w-full rounded-lg border border-gray-300 bg-white px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
              />
            </div>
            <div>
              <label className="text-xs font-semibold text-gray-600">Categoría</label>
              <select
                value={form.category}
                onChange={(e) => setField('category', e.target.value)}
                className="mt-1.5 w-full rounded-lg border border-gray-300 bg-white px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
              >
                <option value="">Sin categoría</option>
                {cats.map((c) => (
                  <option key={c.id} value={c.name}>{c.name}</option>
                ))}
              </select>
            </div>
            <div className="sm:col-span-2">
              <label className="text-xs font-semibold text-gray-600">Descripción</label>
              <textarea
                rows={3}
                value={form.description}
                onChange={(e) => setField('description', e.target.value)}
                placeholder="Describe tu producto, origen, cualidades..."
                className="mt-1.5 w-full rounded-lg border border-gray-300 bg-white px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand resize-none"
              />
            </div>
          </div>
          <div className="flex justify-end gap-3 pt-2">
            <Button type="button" variant="outline" onClick={handleCancel} disabled={saving}>
              Cancelar
            </Button>
            <Button type="submit" variant="primary" disabled={saving}>
              {saving ? (
                <>
                  <Icon name="progress_activity" size={16} className="animate-spin" />
                  Guardando...
                </>
              ) : editingId ? 'Guardar cambios' : 'Publicar producto'}
            </Button>
          </div>
        </form>
      )}

      {offeringsList.length === 0 && !showForm ? (
        <div className="rounded-xl border border-dashed border-gray-300 bg-white p-12 text-center">
          <Icon name="inventory_2" size={48} className="mx-auto text-gray-300" />
          <h2 className="mt-4 text-lg font-semibold text-gray-900">Sin productos</h2>
          <p className="mt-2 text-sm text-gray-500">
            Agrega tu primer producto para que los compradores lo encuentren.
          </p>
        </div>
      ) : (
        <div className="grid gap-5 sm:grid-cols-2 xl:grid-cols-3">
          {offeringsList.map((offering) => (
            <article
              key={offering.id}
              className="group relative flex flex-col overflow-hidden rounded-2xl border border-gray-100 bg-white transition-shadow hover:shadow-lg"
            >
              <div className="relative aspect-[4/3] bg-gray-100">
                {offering.image_url || getProductImage(offering.id) ? (
                  <img src={offering.image_url || getProductImage(offering.id)} alt={offering.name} className="h-full w-full object-cover" />
                ) : (
                  <span className="absolute left-3 top-3 text-4xl">📦</span>
                )}
                <Badge className="absolute right-3 top-3" tone="brand">Activo</Badge>
              </div>
              <div className="flex flex-1 flex-col p-4">
                <h3 className="font-semibold text-gray-900 line-clamp-1">{offering.name}</h3>
                <p className="mt-2 text-lg font-bold text-brand">
                  {formatPrice(offering.price)}
                </p>
                <div className="mt-4 flex items-center gap-2">
                  <Button variant="outline" size="sm" className="flex-1" onClick={() => openEdit(offering)}>
                    Editar
                  </Button>
                  <Button variant="outline" size="sm" className="flex-1 opacity-40 cursor-not-allowed" disabled>
                    Eliminar
                  </Button>
                </div>
              </div>
            </article>
          ))}
        </div>
      )}
    </div>
  )
}
