import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import Icon from '../../components/ui/Icon'
import Badge from '../../components/ui/Badge'
import StatCard from '../../components/StatCard'
import { useAuth } from '../../context/AuthContext'
import { getDisplayName } from '../../lib/user'
import { useOfferings } from '../../hooks/useOfferings'
import { inquiries, reviews } from '../../services/api'
import { getCompanyId } from '../../lib/session'
import { formatPrice } from '../../lib/format'

export default function ProducerHome() {
  const navigate = useNavigate()
  const { user } = useAuth()
  const { offeringsList } = useOfferings(user?.id)
  const [requests, setRequests] = useState([])
  const [producerRating, setProducerRating] = useState(null)

  useEffect(() => {
    const companyId = getCompanyId()
    if (!companyId) {
      setRequests([])
      return
    }
    inquiries.getByCompany(companyId)
      .then((data) => setRequests(Array.isArray(data) ? data : []))
      .catch(() => setRequests([]))
  }, [user?.id])

  useEffect(() => {
    if (!user?.id) return
    reviews.average('user', user.id)
      .then(setProducerRating)
      .catch(() => setProducerRating(null))
  }, [user?.id])

  const inquiryCount = requests.length
  const hasCompany = !!getCompanyId()

  const stats = [
    { icon: 'inventory_2', value: String(offeringsList.length), label: 'Productos activos', tone: 'brand', trend: offeringsList.length > 0 ? `${offeringsList.length} publicados` : 'Sin productos' },
    { icon: 'mail', value: String(inquiryCount), label: 'Solicitudes recibidas', tone: 'amber', trend: inquiryCount > 0 ? 'Pendientes de respuesta' : 'Sin solicitudes' },
    { icon: 'storefront', value: hasCompany ? '1' : '0', label: 'Organización', tone: 'green', trend: hasCompany ? 'Activa' : 'Sin organización' },
    {
      icon: 'star',
      value: producerRating?.count > 0 ? String(producerRating.average) : '—',
      label: 'Valoración',
      tone: 'brand',
      trend: producerRating?.count > 0 ? `${producerRating.count} reseñas` : 'Sin reseñas',
    },
  ]

  return (
    <div className="space-y-8">
      <header>
        <h1 className="text-3xl font-bold text-gray-900">Resumen del negocio</h1>
        <p className="mt-1 text-gray-500">{getDisplayName(user)}</p>
      </header>

      <section aria-label="Estadísticas principales" className="grid gap-5 sm:grid-cols-2 xl:grid-cols-4">
        {stats.map((stat, i) => (
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
          {inquiryCount > 0 && <Badge tone="red">{inquiryCount} nuevas</Badge>}
        </div>
        <div className="mt-3 rounded-xl border border-gray-100 bg-white overflow-x-auto">
          {requests.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-gray-400">
              <Icon name="inbox" size={40} />
              <p className="mt-2 text-sm">Sin solicitudes por ahora</p>
            </div>
          ) : (
            <table className="w-full text-sm min-w-[600px]" role="table">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left font-semibold text-gray-500">Mensaje</th>
                  <th className="px-4 py-3 text-left font-semibold text-gray-500">Producto</th>
                  <th className="px-4 py-3 text-left font-semibold text-gray-500">Estado</th>
                  <th className="px-4 py-3 text-left font-semibold text-gray-500">Tiempo</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {requests.slice(0, 3).map((req) => (
                  <tr key={req.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 text-gray-600 max-w-[260px] truncate">{req.message || '—'}</td>
                    <td className="px-4 py-3 text-gray-600">{req.offering_name || 'Producto'}</td>
                    <td className="px-4 py-3">
                      <Badge
                        tone={req.status === 'pending' ? 'amber' : 'brand'}
                      >
                        {req.status === 'pending' ? 'Nuevo' : 'Respondido'}
                      </Badge>
                    </td>
                    <td className="px-4 py-3 text-gray-500">
                      {new Date(req.created_at).toLocaleDateString('es-NI')}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </section>

      <section aria-label="Mis productos">
        <h2 className="text-lg font-semibold text-gray-900">Mis productos</h2>
        <div className="mt-3 rounded-xl border border-gray-100 bg-white overflow-x-auto">
          {offeringsList.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-gray-400">
              <Icon name="inventory_2" size={40} />
              <p className="mt-2 text-sm">Aún no tienes productos publicados</p>
            </div>
          ) : (
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
                {offeringsList.slice(0, 5).map((offering) => {
                  const desc = offering.description || ''
                  const catMatch = desc.match(/Category:\s*(.+)/)
                  const category = catMatch?.[1] || 'Sin categoría'
                  const unitMatch = desc.match(/Unit:\s*(\S+)/)
                  const unit = unitMatch?.[1] || 'un'
                  return (
                    <tr key={offering.id} className="hover:bg-gray-50">
                      <td className="px-4 py-3 font-medium text-gray-900">{offering.name}</td>
                      <td className="px-4 py-3">
                        <Badge tone="brand">{category}</Badge>
                      </td>
                      <td className="px-4 py-3 text-gray-600">{formatPrice(offering.price)} / {unit}</td>
                      <td className="px-4 py-3">
                        <Badge tone="brand">Disponible</Badge>
                      </td>
                      <td className="px-4 py-3">
                        <button
                          type="button"
                          onClick={() => navigate('/producer/products')}
                          className="text-sm font-semibold text-brand hover:underline"
                        >
                          Ver
                        </button>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          )}
        </div>
      </section>
    </div>
  )
}
