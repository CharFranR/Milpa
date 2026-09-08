import { cn } from '../../lib/cn'
import Icon from '../../components/ui/Icon'
import Badge from '../../components/ui/Badge'
import { growthStats, regionRanking, categoryStats } from '../../mocks/admin'

export default function AdminReports() {
  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold text-gray-900">Reportes y estadísticas</h1>

      <section aria-label="Crecimiento mensual" className="grid gap-5 sm:grid-cols-3">
        {growthStats.map((stat, i) => (
          <article
            key={i}
            className="rounded-xl border border-gray-100 bg-white p-5"
          >
            <p className="text-sm font-medium text-gray-500">{stat.label}</p>
            <div className="mt-2 flex items-baseline gap-2">
              <p className="text-2xl font-bold text-brand">{stat.value}</p>
              <p className="text-sm text-gray-500">{stat.total}</p>
            </div>
            <div className="mt-3 h-2 w-full rounded-full bg-gray-100 overflow-hidden">
              <div
                className="h-full rounded-full bg-brand"
                style={{ width: `${Math.min(stat.progress * 100, 100)}%` }}
              />
            </div>
          </article>
        ))}
      </section>

      <section aria-label="Regiones más activas" className="rounded-xl border border-gray-100 bg-white p-5">
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
                  className="h-full rounded-full bg-brand"
                  style={{ width: `${(reg.count / 420) * 100}%` }}
                />
              </div>
            </div>
          ))}
        </div>
      </section>

      <section aria-label="Productos por categoría" className="rounded-xl border border-gray-100 bg-white p-5">
        <h2 className="text-lg font-semibold text-gray-900">Productos por categoría</h2>
        <div className="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          {categoryStats.map((cat, i) => (
            <article
              key={i}
              className="rounded-lg bg-brand-soft/50 p-4 text-center"
            >
              <p className="text-2xl font-bold text-brand">{cat.count}</p>
              <p className="text-xs text-gray-500">{cat.category}</p>
              <p className="text-xs text-brand">{cat.percent}%</p>
            </article>
          ))}
        </div>
      </section>
    </div>
  )
}