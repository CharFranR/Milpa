import { useState, useEffect, useCallback } from 'react'
import { admin } from '../../services/api'
import { categories } from '../../services/api'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'

export default function AdminConfig() {
  const [cats, setCats] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [showModal, setShowModal] = useState(false)
  const [editing, setEditing] = useState(null)
  const [form, setForm] = useState({ name: '', description: '', is_active: true, default_expiry_days: '' })

  const fetchCategories = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await categories.getAll()
      setCats(data)
    } catch (e) {
      setError('No se pudieron cargar las categorías')
      console.error(e)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchCategories()
  }, [fetchCategories])

  const handleSubmit = async (e) => {
    e.preventDefault()
    const days = form.default_expiry_days
    if (days !== '' && (!Number.isInteger(Number(days)) || Number(days) <= 0)) {
      alert('Los días de vencimiento predeterminados deben ser un número entero mayor a 0')
      return
    }
    try {
      if (editing) {
        await admin.updateCategory(editing.id, form)
      } else {
        await admin.createCategory(form)
      }
      setShowModal(false)
      setEditing(null)
      setForm({ name: '', description: '', is_active: true, default_expiry_days: '' })
      fetchCategories()
    } catch (e) {
      alert('Error al guardar la categoría')
      console.error(e)
    }
  }

  const handleStatusChange = async (id, isActive) => {
    try {
      await admin.setCategoryStatus(id, isActive)
      fetchCategories()
    } catch (e) {
      alert('Error al cambiar el estado')
      console.error(e)
    }
  }

  const openCreate = () => {
    setEditing(null)
    setForm({ name: '', description: '', is_active: true, default_expiry_days: '' })
    setShowModal(true)
  }

  const openEdit = (cat) => {
    setEditing(cat)
    setForm({ name: cat.name, description: cat.description || '', is_active: cat.is_active, default_expiry_days: cat.default_expiry_days ?? '' })
    setShowModal(true)
  }

  if (loading) {
    return (
      <div className="space-y-4">
        <div className="h-8 bg-gray-100 rounded w-1/4 animate-pulse" />
        <div className="h-64 bg-gray-100 rounded animate-pulse" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-600">{error}</p>
        <Button onClick={fetchCategories} className="mt-4">Reintentar</Button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900">Categorías</h1>
        <Button onClick={openCreate}>Nueva categoría</Button>
      </div>

      <div className="rounded-xl border border-gray-100 bg-white overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm" role="table">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Nombre</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Descripción</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Estado</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Acciones</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {cats.length === 0 ? (
                <tr>
                  <td colSpan={4} className="px-4 py-12 text-center text-gray-500">No hay categorías</td>
                </tr>
              ) : (
                cats.map((cat) => (
                  <tr key={cat.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 font-medium text-gray-900">{cat.name}</td>
                    <td className="px-4 py-3 text-gray-500 max-w-xs truncate">{cat.description || '—'}</td>
                    <td className="px-4 py-3">
                      <Badge tone={cat.is_active ? 'brand' : 'amber'} onClick={() => handleStatusChange(cat.id, !cat.is_active)} className="cursor-pointer">
                        {cat.is_active ? 'Activa' : 'Inactiva'}
                      </Badge>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <Button variant="outline" size="sm" onClick={() => openEdit(cat)}>Editar</Button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {showModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-2xl p-6 w-full max-w-md">
            <h2 className="text-xl font-bold text-gray-900 mb-4">{editing ? 'Editar' : 'Nueva'} categoría</h2>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Nombre</label>
                <input
                  type="text"
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  required
                  className="w-full rounded-xl border border-gray-200 bg-white px-4 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Descripción</label>
                <textarea
                  value={form.description}
                  onChange={(e) => setForm({ ...form, description: e.target.value })}
                  rows={3}
                  className="w-full rounded-xl border border-gray-200 bg-white px-4 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
                />
              </div>
              <div>
                <label htmlFor="default_expiry_days" className="block text-sm font-medium text-gray-700 mb-1">Días de vencimiento predeterminados</label>
                <input
                  type="number"
                  id="default_expiry_days"
                  min="1"
                  step="1"
                  value={form.default_expiry_days}
                  onChange={(e) => setForm({ ...form, default_expiry_days: e.target.value })}
                  placeholder="Sin vencimiento predeterminado"
                  className="w-full rounded-xl border border-gray-200 bg-white px-4 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
                />
                <p className="mt-1 text-xs text-gray-500">Opcional. Solo se aplica a los productos nuevos que no tengan una fecha de vencimiento propia.</p>
              </div>
              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="is_active"
                  checked={form.is_active}
                  onChange={(e) => setForm({ ...form, is_active: e.target.checked })}
                  className="h-4 w-4 rounded border-gray-200 text-brand focus:ring-brand"
                />
                <label htmlFor="is_active" className="text-sm text-gray-700">Activa</label>
              </div>
              <div className="flex justify-end gap-2 pt-2">
                <Button type="button" variant="outline" onClick={() => { setShowModal(false); setEditing(null); setForm({ name: '', description: '', is_active: true, default_expiry_days: '' }); }}>Cancelar</Button>
                <Button type="submit">{editing ? 'Guardar' : 'Crear'}</Button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}