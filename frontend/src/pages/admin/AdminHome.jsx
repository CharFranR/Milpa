import { cn } from '../../lib/cn'
import Badge from '../../components/ui/Badge'
import StatCard from '../../components/StatCard'
import { adminStats, recentRegistrations, pendingReports, growthStats, regionRanking, categoryStats } from '../../mocks/admin'

export default function AdminHome() {
  return (
    <div className="space-y-8">
      <header>
        <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
        <p className="mt-1 text-gray-500">Visión general de la plataforma</p>
      </header>

      <section aria-label="KPIs principales" className="grid gap-5 sm:grid-cols-2 xl:grid-cols-4">
        {adminStats.primary.map((stat, i) => (
          <StatCard
            key={stat.label}
            icon={stat.icon}
            value={stat.value}
            label={stat.label}
            index={i}
            tone={stat.tone}
          >
            <div className="mt-2 text-xs font-medium">
              {stat.sub}
            </div>
          </StatCard>
        ))}
      </section>

      <section aria-label="KPIs secundarios" className="grid gap-5 sm:grid-cols-2 xl:grid-cols-4">
        {adminStats.secondary.map((stat, i) => (
          <StatCard
            key={stat.label}
            icon={stat.icon}
            value={stat.value}
            label={stat.label}
            index={i + 4}
            tone={stat.tone}
          />
        ))}
      </section>

      <div className="grid gap-6 lg:grid-cols-2">
        <section aria-label="Registros recientes" className="rounded-xl border border-gray-100 bg-white">
          <div className="flex items-center justify-between border-b border-gray-100 px-5 py-3">
            <h2 className="text-lg font-semibold text-gray-900">Registros recientes</h2>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm" role="table">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left font-semibold text-gray-500">Usuario</th>
                  <th className="px-4 py-3 text-left font-semibold text-gray-500">Tipo · Ubicación</th>
                  <th className="px-4 py-3 text-left font-semibold text-gray-500">Estado</th>
                  <th className="px-4 py-3 text-left font-semibold text-gray-500">Tiempo</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {recentRegistrations.map((reg, i) => (
                  <tr key={i} className="hover:bg-gray-50">
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <span className="inline-flex h-7 w-7 items-center justify-center rounded-full bg-brand-soft text-brand text-xs font-bold">
                          {reg.name.split(' ').map((n) => n[0]).join('').slice(0, 2).toUpperCase()}
                        </span>
                        <span className="font-medium text-gray-900">{reg.name}</span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-gray-500">{reg.type} · {reg.location}</td>
                    <td className="px-4 py-3">
                      <Badge tone={reg.status === 'Activo' ? 'brand' : 'amber'}>
                        {reg.status}
                      </Badge>
                    </td>
                    <td className="px-4 py-3 text-gray-500">{reg.time}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>

        <section aria-label="Reportes pendientes" className="rounded-xl border border-gray-100 bg-white">
          <div className="flex items-center justify-between border-b border-gray-100 px-5 py-3">
            <h2 className="text-lg font-semibold text-gray-900">Reportes pendientes</h2>
          </div>
          <div className="divide-y divide-gray-100 p-3">
            {pendingReports.map((rep, i) => (
              <div key={i} className="flex items-center justify-between gap-3 py-3">
                <div className="flex items-center gap-3">
                  <span
                    className={cn(
                      'inline-flex h-2.5 w-2.5 rounded-full',
                      rep.color === 'red' && 'bg-red-500',
                      rep.color === 'yellow' && 'bg-yellow-500',
                      rep.color === 'gray' && 'bg-gray-400',
                    )}
                  />
                  <span className="text-sm font-medium text-gray-900">{rep.title}</span>
                </div>
                <Badge tone={rep.color === 'red' ? 'red' : rep.color === 'yellow' ? 'amber' : 'gray'}>
                  {rep.severity}
                </Badge>
              </div>
            ))}
          </div>
        </section>
      </div>

      <section aria-label="Crecimiento mensual" className="rounded-xl border border-gray-100 bg-white p-6">
        <h2 className="text-lg font-semibold text-gray-900">Crecimiento mensual</h2>
        <div className="mt-4 space-y-4">
          {growthStats.map((stat, i) => (
            <div key={i} className="space-y-1.5">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-gray-900">{stat.label}</span>
                <span className="text-sm text-gray-500">{stat.value} / {stat.total}</span>
              </div>
              <div className="h-2 w-full rounded-full bg-gray-100 overflow-hidden">
                <div
                  className="h-full rounded-full bg-brand"
                  style={{ width: `${Math.min(stat.progress * 100, 100)}%` }}
                />
              </div>
            </div>
          ))}
        </div>
      </section>

      <div className="grid gap-6 lg:grid-cols-2">
        <section aria-label="Regiones más activas" className="rounded-xl border border-gray-100 bg-white p-6">
          <h2 className="text-lg font-semibold text-gray-900">Regiones más activas</h2>
          <div className="mt-4 space-y-3">
            {regionRanking.map((reg, i) => (
              <div key={i} className="space-y-1.5">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium text-gray-900">{reg.region}</span>
                  <span className="text-sm text-gray-500">{reg.count}</span>
                </div>
                <div className="h-2 w-full rounded-full bg-gray-100 overflow-hidden">
                  <div
                    className="h-full rounded-full bg-brand-soft"
                    style={{ width: `${(reg.count / 420) * 100}%` }}
                  />
                </div>
              </div>
            ))}
          </div>
        </section>

        <section aria-label="Productos por categoría" className="rounded-xl border border-gray-100 bg-white p-6">
          <h2 className="text-lg font-semibold text-gray-900">Productos por categoría</h2>
          <div className="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            {categoryStats.map((cat, i) => (
              <div
                key={i}
                className="rounded-lg bg-brand-soft/50 p-4 text-center"
              >
                <p className="text-2xl font-bold text-brand">{cat.count}</p>
                <p className="text-xs text-gray-500">{cat.category}</p>
                <p className="text-xs text-brand">{cat.percent}%</p>
              </div>
            ))}
          </div>
        </section>
      </div>
    </div>
  )
}