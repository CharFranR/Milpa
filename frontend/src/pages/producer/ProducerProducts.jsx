import { useState, useEffect, useRef } from 'react'
import Icon from '../../components/ui/Icon'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'
import { useOfferings } from '../../hooks/useOfferings'
import { useCompany } from '../../hooks/useCompany'
import { useAuth } from '../../context/AuthContext'
import { categories } from '../../services/api'
import { formatPrice } from '../../lib/format'
import { setProductImage, getProductImage, embedImageInDescription, extractImageFromDescription, resolveOfferingImage } from '../../lib/productImages'
import { compressImage } from '../../lib/imageCompression'
import { Link, useNavigate } from 'react-router-dom'

const MAX_IMAGE_SIZE = 5 * 1024 * 1024
const RENEW_DAYS = 30

// El backend responde en inglés (domain/entities/errors.go); la UI va en español.
const SERVER_ERROR_ES = [
  [/a complete address/i, 'Para publicar necesitas una dirección completa en tu perfil: departamento, municipio y dirección.'],
  [/variety is required/i, 'La variedad es obligatoria.'],
  [/category is required/i, 'Selecciona una categoría.'],
  [/unit of measure is required/i, 'La categoría seleccionada no tiene unidad de medida asociada.'],
  [/quantity must be greater than zero/i, 'La cantidad debe ser mayor que cero.'],
  [/price must be greater than zero/i, 'El precio debe ser mayor que cero.'],
  [/name is required/i, 'El nombre es obligatorio.'],
]

function toSpanish(message = '') {
  const match = SERVER_ERROR_ES.find(([pattern]) => pattern.test(message))
  return match ? match[1] : message
}

function offeringStatus(offering) {
  if (!offering.is_active) return { label: 'Desactivado', tone: 'gray', renew: true }
  const expired = offering.expires_at && new Date(offering.expires_at) <= new Date()
  if (expired) return { label: 'Vencido', tone: 'amber', renew: true }
  return { label: 'Activo', tone: 'brand', renew: false }
}

const EMPTY_FORM = {
  name: '',
  variety: '',
  price: '',
  quantity: '',
  category: '',
  description: '',
  image_url: '',
}

export default function ProducerProducts() {
  const navigate = useNavigate()
  const { user } = useAuth()
  const userId = user?.id
  // La empresa se resuelve por owner: sin esto, un login nuevo no encuentra su
  // empresa hasta que visita "Mi negocio" (getCompanyId() sólo cachea).
  const { company, loading: companyLoading } = useCompany(userId)
  const companyId = company?.id || null
  const { offeringsList, loading, error, createOffering, updateOffering, deactivateOffering, renewOffering } = useOfferings(userId)
  const [showForm, setShowForm] = useState(false)
  const [editingId, setEditingId] = useState(null)
  const [saving, setSaving] = useState(false)
  const [formError, setFormError] = useState('')
  const [actionError, setActionError] = useState('')
  const [busyId, setBusyId] = useState(null)
  const [cats, setCats] = useState([])
  const [form, setForm] = useState(EMPTY_FORM)
  const fileRef = useRef(null)
  const [imagePreview, setImagePreview] = useState('')

  useEffect(() => {
    categories.getAll().then((data) => setCats(Array.isArray(data) ? data : [])).catch(() => {})
  }, [])

  const selectedCategory = cats.find((c) => c.id === form.category)
  const unitOfMeasureId = selectedCategory?.default_unit_of_measure_id || null

  function setField(key, value) {
    setForm((f) => ({ ...f, [key]: value }))
  }

  async function handleImageSelect(e) {
    const file = e.target.files?.[0]
    if (!file) return

    if (file.size > MAX_IMAGE_SIZE) {
      setFormError('La imagen no puede superar 5 MB.')
      return
    }

    try {
      const dataUrl = await compressImage(file)
      setForm((f) => ({ ...f, image_url: dataUrl }))
      setImagePreview(dataUrl)
    } catch {
      setFormError('No se pudo leer la imagen.')
    }
  }

  function handleRemoveImage() {
    setForm((f) => ({ ...f, image_url: '' }))
    setImagePreview('')
    if (fileRef.current) fileRef.current.value = ''
  }

  function openCreate() {
    setForm({ ...EMPTY_FORM })
    setImagePreview('')
    setEditingId(null)
    setShowForm(true)
    setFormError('')
  }

  function openEdit(offering) {
    const { clean: cleanDesc, imageUrl: descImage } = extractImageFromDescription(offering.description || '')
    // Líneas heredadas del formulario viejo: unidad/cantidad/categoría ya viven en columnas.
    const legacyCategory = cleanDesc.match(/Category:\s*(.+)/)?.[1]?.trim()
    const descOnly = cleanDesc
      .replace(/Unit:\s*\S+\n?/g, '')
      .replace(/Qty:\s*\d+\n?/g, '')
      .replace(/Category:\s*.+\n?/g, '')
      .trim()

    const savedImage = descImage || offering.image_url || getProductImage(offering.id) || ''

    setForm({
      name: offering.name || '',
      variety: offering.variety || '',
      price: String(offering.price || ''),
      quantity: offering.quantity_available ? String(offering.quantity_available) : '',
      category: offering.category_id || cats.find((c) => c.name === legacyCategory)?.id || '',
      description: descOnly,
      image_url: savedImage,
    })
    setImagePreview(savedImage)
    setEditingId(offering.id)
    setShowForm(true)
    setFormError('')
  }

  function handleCancel() {
    setForm({ ...EMPTY_FORM })
    setImagePreview('')
    setEditingId(null)
    setShowForm(false)
    setFormError('')
  }

  function validate() {
    if (!form.name.trim()) return 'El nombre es obligatorio.'
    if (!form.variety.trim()) return 'La variedad es obligatoria.'
    if (!form.price || Number(form.price) <= 0) return 'El precio debe ser mayor que cero.'
    if (!form.quantity || Number(form.quantity) <= 0) return 'La cantidad debe ser mayor que cero.'
    if (!form.category) return 'Selecciona una categoría.'
    if (!unitOfMeasureId) return 'La categoría seleccionada no tiene unidad de medida asociada.'
    if (!companyId) return 'Primero debes crear tu empresa en "Mi negocio".'
    return ''
  }

  function handleSubmit(e) {
    e.preventDefault()
    const validationError = validate()
    if (validationError) {
      setFormError(validationError)
      return
    }

    setSaving(true)
    setFormError('')

    const description = embedImageInDescription(form.description.trim(), form.image_url)

    const payload = {
      user_id: userId,
      company_id: companyId,
      type: 0,
      name: form.name.trim(),
      description,
      price: parseFloat(form.price),
      variety: form.variety.trim(),
      category_id: form.category,
      unit_of_measure_id: unitOfMeasureId,
      quantity_available: parseFloat(form.quantity),
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
        setFormError(toSpanish(err.message || 'Error al guardar.'))
      })
      .finally(() => setSaving(false))
  }

  function handleDeactivate(offering) {
    setBusyId(offering.id)
    setActionError('')
    deactivateOffering(offering.id)
      .catch((err) => setActionError(toSpanish(err.message || 'No se pudo desactivar el producto.')))
      .finally(() => setBusyId(null))
  }

  function handleRenew(offering) {
    setBusyId(offering.id)
    setActionError('')
    const expiresAt = new Date(Date.now() + RENEW_DAYS * 86400000).toISOString()
    renewOffering(offering.id, expiresAt)
      .catch((err) => setActionError(toSpanish(err.message || 'No se pudo renovar el producto.')))
      .finally(() => setBusyId(null))
  }

  if (loading || companyLoading) {
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
          <Button type="button" variant="primary" className="mt-6" onClick={() => navigate('/producer/business')} icon={<Icon name="storefront" size={18} />}>
            Ir a Mi negocio
          </Button>
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

      {actionError && (
        <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{actionError}</p>
      )}

      {showForm && (
        <form onSubmit={handleSubmit} className="rounded-xl border border-gray-200 bg-brand-soft/50 p-6 space-y-4">
          <h2 className="text-lg font-semibold text-gray-900">
            {editingId ? 'Editar producto' : 'Nuevo producto'}
          </h2>

          {formError && (
            <div className="rounded-lg bg-red-50 px-3 py-2 text-xs text-red-700">
              <p>{formError}</p>
              {/dirección completa/i.test(formError) && (
                <Link to="/producer/business" className="mt-1 inline-block font-semibold underline">Completar dirección en Mi negocio</Link>
              )}
            </div>
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
              <label className="text-xs font-semibold text-gray-600">Variedad *</label>
              <input
                type="text"
                required
                value={form.variety}
                onChange={(e) => setField('variety', e.target.value)}
                placeholder="Cherry, Criollo, Híbrido..."
                className="mt-1.5 w-full rounded-lg border border-gray-300 bg-white px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
              />
            </div>

            <div>
              <label className="text-xs font-semibold text-gray-600">Precio (C$) *</label>
              <input
                type="number"
                required
                min="1"
                value={form.price}
                onChange={(e) => setField('price', e.target.value)}
                placeholder="180"
                className="mt-1.5 w-full rounded-lg border border-gray-300 bg-white px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
              />
            </div>

            <div>
              <label className="text-xs font-semibold text-gray-600">Cantidad disponible *</label>
              <input
                type="number"
                required
                min="1"
                value={form.quantity}
                onChange={(e) => setField('quantity', e.target.value)}
                placeholder="100"
                className="mt-1.5 w-full rounded-lg border border-gray-300 bg-white px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
              />
            </div>

            <div>
              <label className="text-xs font-semibold text-gray-600">Categoría *</label>
              <select
                required
                value={form.category}
                onChange={(e) => setField('category', e.target.value)}
                className="mt-1.5 w-full rounded-lg border border-gray-300 bg-white px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
              >
                <option value="">Selecciona una categoría</option>
                {cats.map((c) => (
                  <option key={c.id} value={c.id}>{c.name}</option>
                ))}
              </select>
            </div>

            <div>
              <label className="text-xs font-semibold text-gray-600">Unidad de medida</label>
              <p className="mt-1.5 rounded-lg border border-dashed border-gray-300 bg-white px-3 py-2.5 text-sm text-gray-500">
                {form.category
                  ? unitOfMeasureId
                    ? 'Fijada por la categoría seleccionada'
                    : 'La categoría no tiene unidad asociada'
                  : 'Elige una categoría'}
              </p>
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
          {offeringsList.map((offering) => {
            const status = offeringStatus(offering)
            const busy = busyId === offering.id
            return (
              <article
                key={offering.id}
                className="group relative flex flex-col overflow-hidden rounded-2xl border border-gray-100 bg-white transition-shadow hover:shadow-lg"
              >
                <div className="relative aspect-[4/3] bg-gray-100">
                  {resolveOfferingImage(offering) ? (
                    <img src={resolveOfferingImage(offering)} alt={offering.name} className="h-full w-full object-cover" />
                  ) : (
                    <span className="absolute left-3 top-3 text-4xl">📦</span>
                  )}
                  <Badge className="absolute right-3 top-3" tone={status.tone}>{status.label}</Badge>
                </div>
                <div className="flex flex-1 flex-col p-4">
                  <h3 className="font-semibold text-gray-900 line-clamp-1">{offering.name}</h3>
                  <p className="mt-1 text-xs text-gray-500">
                    {offering.variety || 'Sin variedad'} · {offering.quantity_available ?? 0} disponibles
                  </p>
                  <p className="mt-2 text-lg font-bold text-brand">
                    {formatPrice(offering.price)}
                  </p>
                  <div className="mt-4 flex items-center gap-2">
                    <Button variant="outline" size="sm" className="flex-1" onClick={() => openEdit(offering)} disabled={busy}>
                      Editar
                    </Button>
                    {status.renew ? (
                      <Button variant="primary" size="sm" className="flex-1" onClick={() => handleRenew(offering)} disabled={busy}>
                        Renovar
                      </Button>
                    ) : (
                      <Button variant="outline" size="sm" className="flex-1" onClick={() => handleDeactivate(offering)} disabled={busy}>
                        Desactivar
                      </Button>
                    )}
                  </div>
                </div>
              </article>
            )
          })}
        </div>
      )}
    </div>
  )
}
