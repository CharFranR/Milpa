import { useState, useEffect, useCallback } from 'react'
import { offerings, admin } from '../../services/api'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'

export default function AdminProducts() {
  const [products, setProducts] = useState([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize] = useState(20)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  const fetchProducts = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await offerings.search({
        page,
        page_size: pageSize,
        sort: 'relevance',
      })
      setProducts(data.results || [])
      setTotal(data.total_hits || data.total || 0)
    } catch (e) {
      setError('No se pudieron cargar los productos')
      console.error(e)
    } finally {
      setLoading(false)
    }
  }, [page, pageSize])

  useEffect(() => {
    fetchProducts()
  }, [fetchProducts])

  const handleDelete = async (id) => {
    if (!window.confirm('¿Eliminar este producto definitivamente?')) return
    try {
      await admin.deleteOffering(id)
      fetchProducts()
    } catch (e) {
      alert('Error al eliminar el producto')
      console.error(e)
    }
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
        <Button onClick={fetchProducts} className="mt-4">Reintentar</Button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900">Productos</h1>
      </div>

      <div className="rounded-xl border border-gray-100 bg-white overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm" role="table">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Producto</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Productor</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Categoría</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Precio</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Estado</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Acciones</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {products.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-4 py-12 text-center text-gray-500">No hay productos</td>
                </tr>
              ) : (
                products.map((prod) => (
                  <tr key={prod.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-3">
                        <span className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-gray-100 text-gray-400 text-2xl">📦</span>
                        <span className="font-medium text-gray-900">{prod.name}</span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-gray-600">{prod.company_name || prod.farmer_name || 'Desconocido'}</td>
                    <td className="px-4 py-3">
                      <Badge tone="brand">{prod.category_name || prod.category_id || 'Sin categoría'}</Badge>
                    </td>
                    <td className="px-4 py-3 text-gray-600">
                      {prod.price ? `${prod.price.toLocaleString()} C$` : 'N/A'}
                    </td>
                    <td className="px-4 py-3">
                      <Badge tone={prod.is_active ? 'brand' : 'amber'}>
                        {prod.is_active ? 'Disponible' : 'Inactivo'}
                      </Badge>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <Button variant="outline" size="sm" onClick={() => {}}>Ver</Button>
                        <Button variant="outline" size="sm" className="text-red-600 hover:bg-red-50" onClick={() => handleDelete(prod.id)}>
                          Eliminar
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
        {total > pageSize && (
          <div className="px-4 py-3 border-t border-gray-100 flex items-center justify-between">
            <span className="text-sm text-gray-500">
              Mostrando {(page - 1) * pageSize + 1}–{Math.min(page * pageSize, total)} de {total}
            </span>
            <div className="flex gap-2">
              <Button variant="outline" size="sm" onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page === 1}>Anterior</Button>
              <Button variant="outline" size="sm" onClick={() => setPage(p => p + 1)} disabled={page * pageSize >= total}>Siguiente</Button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}