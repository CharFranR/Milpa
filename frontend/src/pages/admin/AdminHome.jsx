import { useState, useEffect, useCallback } from 'react'
import { admin, reports, offerings } from '../../services/api'
import { cn } from '../../lib/cn'
import Badge from '../../components/ui/Badge'
import StatCard from '../../components/StatCard'

export default function AdminHome() {
  const [stats, setStats] = useState({
    totalUsers: { value: '—', loading: true },
    totalProducers: { value: '—', loading: true },
    totalProducts: { value: '—', loading: true },
    pendingReports: { value: '—', loading: true },
  })
  const [recentUsers, setRecentUsers] = useState([])
  const [pendingReportsList, setPendingReportsList] = useState([])
  const [loading, setLoading] = useState(true)

  const fetchAll = useCallback(async () => {
    setLoading(true)
    try {
      const [usersRes, productsRes, reportsRes] = await Promise.allSettled([
        admin.listUsers({ page: 1, page_size: 1 }),
        offerings.search({ page: 1, page_size: 1, sort: 'relevance' }),
        reports.list({ status: 'pending', page: 1, page_size: 10 }),
      ])

      const totalUsers = usersRes.status === 'fulfilled' ? usersRes.value.total : 0
      const totalProducts = productsRes.status === 'fulfilled' ? (productsRes.value.total_hits || productsRes.value.total || 0) : 0
      const pendingReportsCount = reportsRes.status === 'fulfilled' ? reportsRes.value.total : 0

      // Get producers count from users (role 1 = producer)
      let totalProducers = 0
      if (usersRes.status === 'fulfilled' && usersRes.value.items) {
        const { normalizeRole } = await import('../../lib/roles')
        totalProducers = usersRes.value.items.filter(u => normalizeRole(u.role) === 'producer').length
      }

      setStats({
        totalUsers: { value: totalUsers.toLocaleString(), loading: false },
        totalProducers: { value: totalProducers.toLocaleString(), loading: false },
        totalProducts: { value: totalProducts.toLocaleString(), loading: false },
        pendingReports: { value: pendingReportsCount.toLocaleString(), loading: false },
      })

      // Fetch recent users (first page)
      const recentRes = await admin.listUsers({ page: 1, page_size: 5 })
      setRecentUsers(recentRes.items || [])

      // Fetch pending reports details
      if (reportsRes.status === 'fulfilled') {
        setPendingReportsList(reportsRes.items || [])
      }
    } catch (e) {
      console.error('Error fetching admin home stats:', e)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchAll()
  }, [fetchAll])

  const primaryStats = [
    { label: 'Usuarios totales', value: stats.totalUsers.value, sub: stats.totalUsers.loading ? 'Cargando...' : '↑ Datos reales', tone: 'blue', icon: 'group' },
    { label: 'Productores activos', value: stats.totalProducers.value, sub: stats.totalProducers.loading ? 'Cargando...' : '↑ Datos reales', tone: 'brand', icon: 'agriculture' },
    { label: 'Productos publicados', value: stats.totalProducts.value, sub: stats.totalProducts.loading ? 'Cargando...' : '↑ Datos reales', tone: 'amber', icon: 'inventory_2' },
    { label: 'Reportes pendientes', value: stats.pendingReports.value, sub: stats.pendingReports.loading ? 'Cargando...' : '↑ Datos reales', tone: 'red', icon: 'flag' },
  ]

  const secondaryStats = [
    { label: 'Municipios activos', value: '—', icon: 'map', tone: 'brand', demo: true },
    { label: 'Mensajes enviados', value: '—', icon: 'chat_bubble', tone: 'blue', demo: true },
    { label: 'Nuevos hoy', value: '—', icon: 'person_add', tone: 'amber', demo: true },
    { label: 'Valoración media', value: '—', icon: 'star', tone: 'red', demo: true },
  ]

  if (loading) {
    return (
      <div className="space-y-8">
        <header>
          <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
          <p className="mt-1 text-gray-500">Visión general de la plataforma</p>
        </header>
        <section aria-label="KPIs principales" className="grid gap-5 sm:grid-cols-2 xl:grid-cols-4">
          {primaryStats.map((stat, i) => (
            <StatCard key={stat.label} icon={stat.icon} value="—" label={stat.label} index={i} tone={stat.tone}>
              <div className="mt-2 text-xs font-medium animate-pulse">Cargando...</div>
            </StatCard>
          ))}
        </section>
      </div>
    )
  }

  return (
    <div className="space-y-8">
      <header>
        <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
        <p className="mt-1 text-gray-500">Visión general de la plataforma</p>
      </header>

      <section aria-label="KPIs principales" className="grid gap-5 sm:grid-cols-2 xl:grid-cols-4">
        {primaryStats.map((stat, i) => (
          <StatCard
            key={stat.label}
            icon={stat.icon}
            value={stat.value}
            label={stat.label}
            index={i}
            tone={stat.tone}
          >
            <div className="mt-2 text-xs font-medium text-green-600">
              {stat.sub}
            </div>
          </StatCard>
        ))}
      </section>

      <section aria-label="KPIs secundarios" className="grid gap-5 sm:grid-cols-2 xl:grid-cols-4">
        {secondaryStats.map((stat, i) => (
          <StatCard
            key={stat.label}
            icon={stat.icon}
            value={stat.value}
            label={stat.label}
            index={i + 4}
            tone={stat.tone}
          >
            {stat.demo && <Badge tone="amber" className="mt-2 text-xs">Demo</Badge>}
          </StatCard>
        ))}
      </section>

      <div className="grid gap-6 lg:grid-cols-2">
        <section aria-label="Registros recientes" className="rounded-xl border border-gray-100 bg-white">
          <div className="flex items-center justify-between border-b border-gray-100 px-5 py-3">
            <h2 className="text-lg font-semibold text-gray-900">Registros recientes</h2>
            <Badge tone="blue" className="text-xs">Real</Badge>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm" role="table">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left font-semibold text-gray-500">Usuario</th>
                  <th className="px-4 py-3 text-left font-semibold text-gray-500">Tipo · Ubicación</th>
                  <th className="px-4 py-3 text-left font-semibold text-gray-500">Estado</th>
                  <th className="px-4 py-3 text-left font-semibold text-gray-500">Fecha</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {recentUsers.length === 0 ? (
                  <tr>
                    <td colSpan={4} className="px-4 py-12 text-center text-gray-500">No hay usuarios recientes</td>
                  </tr>
                ) : (
                  recentUsers.map((user, i) => {
                    const { normalizeRole, ROLE_LABELS } = require('../../lib/roles')
                    const role = normalizeRole(user.role)
                    const label = ROLE_LABELS[role] || 'Desconocido'
                    return (
                      <tr key={user.id || i} className="hover:bg-gray-50">
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-2">
                            <span className="inline-flex h-7 w-7 items-center justify-center rounded-full bg-brand-soft text-brand text-xs font-bold">
                              {(user.first_name?.[0] || '') + (user.last_name?.[0] || '')}
                            </span>
                            <span className="font-medium text-gray-900">{user.first_name} {user.last_name}</span>
                          </div>
                        </td>
                        <td className="px-4 py-3 text-gray-500">{label} · {user.municipality} / {user.department}</td>
                        <td className="px-4 py-3">
                          <Badge tone={user.suspended_at ? 'amber' : role === 'pending' ? 'amber' : 'brand'}>
                            {user.suspended_at ? 'Suspendido' : role === 'pending' ? 'Pendiente' : 'Activo'}
                          </Badge>
                        </td>
                        <td className="px-4 py-3 text-gray-500">
                          {new Date(user.created_at).toLocaleDateString('es-ES', { year: 'numeric', month: 'short', day: 'numeric' })}
                        </td>
                      </tr>
                    )
                  }
                )
              )}
              </tbody>
            </table>
          </div>
        </section>

        <section aria-label="Reportes pendientes" className="rounded-xl border border-gray-100 bg-white">
          <div className="flex items-center justify-between border-b border-gray-100 px-5 py-3">
            <h2 className="text-lg font-semibold text-gray-900">Reportes pendientes</h2>
            <Badge tone="blue" className="text-xs">Real</Badge>
          </div>
          <div className="divide-y divide-gray-100 p-3">
            {pendingReportsList.length === 0 ? (
              <div className="py-6 text-center text-gray-500">No hay reportes pendientes</div>
            ) : (
              pendingReportsList.map((rep, i) => {
                const tones = { Alta: 'red', Media: 'amber', Baja: 'gray' }
                const severityTone = rep.severity ? tones[rep.severity] || 'gray' : 'gray'
                return (
                  <div key={rep.id || i} className="flex items-center justify-between gap-3 py-3">
                    <div className="flex items-center gap-3">
                      <span className={cn('inline-flex h-2.5 w-2.5 rounded-full', severityTone === 'red' && 'bg-red-500', severityTone === 'amber' && 'bg-yellow-500', severityTone === 'gray' && 'bg-gray-400')} />
                      <span className="text-sm font-medium text-gray-900">{rep.title || 'Sin título'}</span>
                    </div>
                    <Badge tone={severityTone}>{rep.severity || '—'}</Badge>
                  </div>
                )
              }
            )
            )}
          </div>
        </section>
      </div>

      <section aria-label="Crecimiento mensual" className="rounded-xl border border-gray-100 bg-white p-6 relative">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">Crecimiento mensual</h2>
          <Badge tone="amber" className="text-xs">Demo</Badge>
        </div>
        <div className="mt-4 space-y-4">
          <div className="space-y-1.5">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium text-gray-900">Nuevos usuarios</span>
              <span className="text-sm text-gray-500">—</span>
            </div>
            <div className="h-2 w-full rounded-full bg-gray-100 overflow-hidden">
              <div className="h-full rounded-full bg-brand" style={{ width: '0%' }} />
            </div>
          </div>
          <div className="space-y-1.5">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium text-gray-900">Nuevos productores</span>
              <span className="text-sm text-gray-500">—</span>
            </div>
            <div className="h-2 w-full rounded-full bg-gray-100 overflow-hidden">
              <div className="h-full rounded-full bg-brand" style={{ width: '0%' }} />
            </div>
          </div>
          <div className="space-y-1.5">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium text-gray-900">Nuevos productos</span>
              <span className="text-sm text-gray-500">—</span>
            </div>
            <div className="h-2 w-full rounded-full bg-gray-100 overflow-hidden">
              <div className="h-full rounded-full bg-brand" style={{ width: '0%' }} />
            </div>
          </div>
        </div>
      </section>

      <div className="grid gap-6 lg:grid-cols-2">
        <section aria-label="Regiones más activas" className="rounded-xl border border-gray-100 bg-white p-6 relative">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-gray-900">Regiones más activas</h2>
            <Badge tone="amber" className="text-xs">Demo</Badge>
          </div>
          <div className="mt-4 space-y-3">
            <div className="space-y-1.5">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-gray-900">Matagalpa</span>
                <span className="text-sm text-gray-500">—</span>
              </div>
              <div className="h-2 w-full rounded-full bg-gray-100 overflow-hidden">
                <div className="h-full rounded-full bg-brand" style={{ width: '0%' }} />
              </div>
            </div>
            <div className="space-y-1.5">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-gray-900">Jinotega</span>
                <span className="text-sm text-gray-500">—</span>
              </div>
              <div className="h-2 w-full rounded-full bg-gray-100 overflow-hidden">
                <div className="h-full rounded-full bg-brand" style={{ width: '0%' }} />
              </div>
            </div>
            <div className="space-y-1.5">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-gray-900">Nueva Segovia</span>
                <span className="text-sm text-gray-500">—</span>
              </div>
              <div className="h-2 w-full rounded-full bg-gray-100 overflow-hidden">
                <div className="h-full rounded-full bg-brand" style={{ width: '0%' }} />
              </div>
            </div>
          </div>
        </section>

        <section aria-label="Productos por categoría" className="rounded-xl border border-gray-100 bg-white p-6 relative">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-gray-900">Productos por categoría</h2>
            <Badge tone="amber" className="text-xs">Demo</Badge>
          </div>
          <div className="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            {['Granos Básicos', 'Café', 'Frutas Tropicales', 'Lácteos y Quesos', 'Hortalizas', 'Carnes y Ganadería', 'Plátanos y Cocos', 'Miel y Apicultura'].map((cat) => (
              <div key={cat} className="rounded-lg bg-brand-soft/50 p-4 text-center">
                <p className="text-2xl font-bold text-brand">—</p>
                <p className="text-xs text-gray-500">{cat}</p>
                <p className="text-xs text-brand">—%</p>
              </div>
            ))}
          </div>
        </section>
      </div>
    </div>
  )
}