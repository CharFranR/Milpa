import { useState, useEffect, useCallback } from 'react'
import { admin, reports, offerings } from '../../services/api'
import { cn } from '../../lib/cn'
import { normalizeRole, ROLE_LABELS } from '../../lib/roles'
import Badge from '../../components/ui/Badge'
import StatCard from '../../components/StatCard'

const DAY_MS = 24 * 60 * 60 * 1000

function BarsList({ items, caption }) {
  if (items.length === 0) {
    return <div className="py-6 text-center text-sm text-gray-500">Sin datos aún</div>
  }
  return (
    <>
      <div className="mt-4 space-y-3">
        {items.map((it) => (
          <div key={it.name} className="space-y-1.5">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium text-gray-900">{it.name}</span>
              <span className="text-sm text-gray-500">{it.count}</span>
            </div>
            <div className="h-2 w-full rounded-full bg-gray-100 overflow-hidden">
              <div
                className="h-full rounded-full bg-brand"
                style={{ width: `${(it.count / items[0].count) * 100}%` }}
              />
            </div>
          </div>
        ))}
      </div>
      {caption && <p className="mt-4 text-xs text-gray-400">{caption}</p>}
    </>
  )
}

export default function AdminHome() {
  const [stats, setStats] = useState({
    totalUsers: '—',
    totalProducers: '—',
    totalProducts: '—',
    pendingReports: '—',
    newToday: '—',
    new30d: '—',
    municipalities: '—',
    departments: '—',
  })
  const [recentUsers, setRecentUsers] = useState([])
  const [pendingReportsList, setPendingReportsList] = useState([])
  const [regions, setRegions] = useState(null)
  const [municipalities, setMunicipalities] = useState(null)
  const [searchCovered, setSearchCovered] = useState(0)
  const [searchTotal, setSearchTotal] = useState(0)
  const [loading, setLoading] = useState(true)

  const fetchAll = useCallback(async () => {
    setLoading(true)
    try {
      const [usersRes, productsRes, reportsRes, searchRes] = await Promise.allSettled([
        admin.listUsers({ page: 1, pageSize: 100 }),
        offerings.search({ page: 1, page_size: 1 }),
        reports.list({ status: 'pending', page: 1, pageSize: 10 }),
        offerings.search({ page: 1, page_size: 100 }),
      ])

      const allUsers = []
      let totalUsers = 0
      if (usersRes.status === 'fulfilled') {
        totalUsers = usersRes.value.total || 0
        allUsers.push(...(usersRes.value.items || []))
        let page = 2
        while (allUsers.length < totalUsers && page <= 20) {
          const next = await admin.listUsers({ page, pageSize: 100 }).catch(() => null)
          if (!next || !next.items || next.items.length === 0) break
          allUsers.push(...next.items)
          page += 1
        }
      }

      const totalProducts = productsRes.status === 'fulfilled' ? (productsRes.value.total_hits || productsRes.value.total || 0) : 0
      const pendingReportsCount = reportsRes.status === 'fulfilled' ? reportsRes.value.total : 0

      const producers = allUsers.filter((u) => normalizeRole(u.role) === 'producer')
      const now = Date.now()
      const newToday = allUsers.filter((u) => u.created_at && now - new Date(u.created_at).getTime() < DAY_MS).length
      const new30d = allUsers.filter((u) => u.created_at && now - new Date(u.created_at).getTime() < 30 * DAY_MS).length

      const results = searchRes.status === 'fulfilled' ? searchRes.value.results || [] : []
      const byDept = new Map()
      const byMun = new Map()
      for (const item of results) {
        const dept = item.department || 'Sin departamento'
        byDept.set(dept, (byDept.get(dept) || 0) + 1)
        const key = item.municipality ? `${item.municipality}, ${dept}` : dept
        byMun.set(key, (byMun.get(key) || 0) + 1)
      }
      const toList = (m) => [...m.entries()].map(([name, count]) => ({ name, count })).sort((a, b) => b.count - a.count)
      const regionList = toList(byDept).slice(0, 5)
      const munList = toList(byMun).slice(0, 5)

      setStats({
        totalUsers: totalUsers.toLocaleString(),
        totalProducers: producers.length.toLocaleString(),
        totalProducts: totalProducts.toLocaleString(),
        pendingReports: pendingReportsCount.toLocaleString(),
        newToday: newToday.toLocaleString(),
        new30d: new30d.toLocaleString(),
        municipalities: byMun.size.toLocaleString(),
        departments: byDept.size.toLocaleString(),
      })
      setRegions(regionList)
      setMunicipalities(munList)
      setSearchCovered(results.length)
      setSearchTotal(searchRes.status === 'fulfilled' ? searchRes.value.total_hits || results.length : 0)

      setRecentUsers(allUsers.slice(0, 8))
      if (reportsRes.status === 'fulfilled') {
        setPendingReportsList(reportsRes.value.items || [])
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
    { label: 'Usuarios totales', value: stats.totalUsers, tone: 'blue', icon: 'group' },
    { label: 'Productores activos', value: stats.totalProducers, tone: 'brand', icon: 'agriculture' },
    { label: 'Productos publicados', value: stats.totalProducts, tone: 'amber', icon: 'inventory_2' },
    { label: 'Reportes pendientes', value: stats.pendingReports, tone: 'red', icon: 'flag' },
  ]

  const secondaryStats = [
    { label: 'Nuevos hoy', value: stats.newToday, icon: 'person_add', tone: 'amber' },
    { label: 'Nuevos (30 días)', value: stats.new30d, icon: 'group', tone: 'blue' },
    { label: 'Municipios con oferta', value: stats.municipalities, icon: 'map', tone: 'brand' },
    { label: 'Departamentos con oferta', value: stats.departments, icon: 'storefront', tone: 'red' },
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
            <StatCard key={stat.label} icon={stat.icon} value="—" label={stat.label} index={i} tone={stat.tone} />
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
          <StatCard key={stat.label} icon={stat.icon} value={stat.value} label={stat.label} index={i} tone={stat.tone} />
        ))}
      </section>

      <section aria-label="KPIs secundarios" className="grid gap-5 sm:grid-cols-2 xl:grid-cols-4">
        {secondaryStats.map((stat, i) => (
          <StatCard key={stat.label} icon={stat.icon} value={stat.value} label={stat.label} index={i + 4} tone={stat.tone} />
        ))}
      </section>

      <div className="grid gap-6 lg:grid-cols-2">
        <section aria-label="Registros recientes" className="min-w-0 rounded-xl border border-gray-100 bg-white">
          <div className="border-b border-gray-100 px-5 py-3">
            <h2 className="text-lg font-semibold text-gray-900">Registros recientes</h2>
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

        <section aria-label="Reportes pendientes" className="min-w-0 rounded-xl border border-gray-100 bg-white">
          <div className="border-b border-gray-100 px-5 py-3">
            <h2 className="text-lg font-semibold text-gray-900">Reportes pendientes</h2>
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

      <div className="grid gap-6 lg:grid-cols-2">
        <section aria-label="Regiones más activas" className="min-w-0 rounded-xl border border-gray-100 bg-white p-6">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-gray-900">Regiones más activas</h2>
            {searchCovered < searchTotal && (
              <span className="text-xs text-gray-400">Primeros {searchCovered} de {searchTotal} anuncios</span>
            )}
          </div>
          <BarsList items={regions || []} />
        </section>

        <section aria-label="Municipios con más oferta" className="min-w-0 rounded-xl border border-gray-100 bg-white p-6">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-gray-900">Municipios con más oferta</h2>
            {searchCovered < searchTotal && (
              <span className="text-xs text-gray-400">Primeros {searchCovered} de {searchTotal} anuncios</span>
            )}
          </div>
          <BarsList items={municipalities || []} />
        </section>
      </div>
    </div>
  )
}
