import { cn } from '../../lib/cn'
import Icon from '../../components/ui/Icon'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'
import { adminProducts } from '../../mocks/admin'

export default function AdminProducts() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900">Productos</h1>
        <Badge tone="amber" className="text-sm">
          7 requieren revisión
        </Badge>
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
              {adminProducts.map((prod, i) => (
                <tr key={i} className="hover:bg-gray-50">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-3">
                      <span className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-gray-100 text-gray-400 text-2xl">📦</span>
                      <span className="font-medium text-gray-900">{prod.name}</span>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-gray-600">{prod.producer}</td>
                  <td className="px-4 py-3">
                    <Badge tone="brand">{prod.category}</Badge>
                  </td>
                  <td className="px-4 py-3 text-gray-600">${prod.price.toLocaleString()}</td>
                  <td className="px-4 py-3">
                    <Badge tone={prod.status === 'Disponible' ? 'brand' : 'amber'}>
                      {prod.status}
                    </Badge>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <Button variant="outline" size="sm">Ver</Button>
                      <Button variant="outline" size="sm" className="text-red-600 hover:bg-red-50">Eliminar</Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}