import { cn } from '../../lib/cn'
import Icon from '../../components/ui/Icon'
import Badge from '../../components/ui/Badge'
import StatCard from '../../components/StatCard'
import { producerStats, producerProducts, producerRequests } from '../../mocks/producer'

export default function ProducerHome() {
  return (
    <div className="space-y-8">
      <header>
        <h1 className="text-3xl font-bold text-gray-900">Resumen del negocio</h1>
        <p className="mt-1 text-gray-500">Finca La Esperanza · Choachí, Cundinamarca</p>
      </header>

      <section aria-label="Estadísticas principales" className="grid gap-5 sm:grid-cols-2 xl:grid-cols-4">
        {producerStats.map((stat, i) => (
          <StatCard
            key={stat.label}
            icon={stat.icon}
            value={stat.value}
            label={stat.label}
            index={i}
            tone={stat.tone}
          >
            <div className="mt-2 flex items-center gap-1.5 text-xs font-medium">
              <Icon name="trending_up" size={12} className="text-green-600" />
              <span className="text-green-600">{stat.trend}</span>
            </div>
          </StatCard>
        ))}
      </section>

      <section aria-label="Solicitudes recientes">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">Solicitudes recientes</h2>
          <Badge tone="red">5 nuevas</Badge>
        </div>
        <div className="mt-3 rounded-xl border border-gray-100 bg-white overflow-x-auto">
          <table className="w-full text-sm min-w-[600px]" role="table">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Comprador</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Producto</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Cantidad</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Estado</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Tiempo</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {producerRequests.slice(0, 3).map((req) => (
                <tr key={req.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <span
                        className={cn(
                          'inline-flex h-7 w-7 items-center justify-center rounded-full text-xs font-bold',
                          req.buyer.tone === 'red' && 'bg-red-100 text-red-600',
                          req.buyer.tone === 'brand' && 'bg-brand-soft text-brand',
                          req.buyer.tone === 'amber' && 'bg-amber-100 text-amber-600',
                          req.buyer.tone === 'blue' && 'bg-blue-100 text-blue-600',
                        )}
                      >
                        {req.buyer.avatar}
                      </span>
                      <span className="font-medium text-gray-900">{req.buyer.name}</span>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-gray-600">{req.product}</td>
                  <td className="px-4 py-3 text-gray-600">{req.qty}</td>
                  <td className="px-4 py-3">
                    <Badge
                      tone={req.status === 'pending' ? 'amber' : 'brand'}
                    >
                      {req.status === 'pending' ? 'Nuevo' : 'Respondido'}
                    </Badge>
                  </td>
                  <td className="px-4 py-3 text-gray-500">{req.time}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      <section aria-label="Mis productos">
        <h2 className="text-lg font-semibold text-gray-900">Mis productos</h2>
        <div className="mt-3 rounded-xl border border-gray-100 bg-white overflow-x-auto">
          <table className="w-full text-sm min-w-[600px]" role="table">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Producto</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Categoría</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Precio</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Estado</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Acciones</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {producerProducts.slice(0, 5).map((prod) => (
                <tr key={prod.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 font-medium text-gray-900">{prod.name}</td>
                  <td className="px-4 py-3">
                    <Badge tone="brand">{prod.category}</Badge>
                  </td>
                  <td className="px-4 py-3 text-gray-600">${prod.price.toLocaleString()} / {prod.unit}</td>
                  <td className="px-4 py-3">
                    <Badge tone={prod.available ? 'brand' : 'amber'}>
                      {prod.available ? 'Disponible' : 'Agotado'}
                    </Badge>
                  </td>
                  <td className="px-4 py-3">
                    <button
                      type="button"
                      className="text-sm font-semibold text-brand hover:underline"
                    >
                      Ver
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>
    </div>
  )
}