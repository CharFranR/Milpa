import { useState, useEffect, useCallback } from 'react'
import { admin } from '../../services/api'
import { growthStats, regionRanking, categoryStats } from '../../mocks/admin'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'

const ACTION_LABELS = {
  user_suspended: 'Usuario suspendido',
  user_reactivated: 'Usuario reactivado',
  user_role_updated: 'Rol de usuario actualizado',
  offering_deleted: 'Producto eliminado',
  report_created: 'Reporte creado',
  report_approved: 'Reporte aprobado',
  report_rejected: 'Reporte rechazado',
}

const TARGET_TYPE_LABELS = {
  user: 'Usuario',
  offering: 'Producto',
  report: 'Reporte',
}

export default function AdminReports() {
  const [auditLogs, setAuditLogs] = useState([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize] = useState(20)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [actionFilter, setActionFilter] = useState('all')
  const [targetTypeFilter, setTargetTypeFilter] = useState('all')

  const fetchAuditLogs = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const params = { page, page_size: pageSize }
      if (actionFilter !== 'all') params.action = actionFilter
      if (targetTypeFilter !== 'all') params.targetType = targetTypeFilter
      const data = await admin.listAuditLogs(params)
      setAuditLogs(data.items || [])
      setTotal(data.total || 0)
    } catch (e) {
      setError('No se pudieron cargar los logs de auditoría')
      console.error(e)
    } finally {
      setLoading(false)
    }
  }, [page, pageSize, actionFilter, targetTypeFilter])

  useEffect(() => {
    fetchAuditLogs()
  }, [fetchAuditLogs])

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
        <Button onClick={fetchAuditLogs} className="mt-4">Reintentar</Button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold text-gray-900">Reportes y estadísticas</h1>

      <section aria-label="KPIs del sistema" className="grid gap-5 sm:grid-cols-3">
        {growthStats.map((stat, i) => (
          <article
            key={i}
            className="rounded-xl border border-gray-100 bg-white p-5 relative"
          >
            <Badge tone="amber" className="absolute top-3 right-3 text-xs">Demo</Badge>
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

      <section aria-label="Regiones más activas" className="rounded-xl border border-gray-100 bg-white p-5 relative">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">Regiones más activas</h2>
          <Badge tone="amber" className="text-xs">Demo</Badge>
        </div>
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

      <section aria-label="Productos por categoría" className="rounded-xl border border-gray-100 bg-white p-5 relative">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">Productos por categoría</h2>
          <Badge tone="amber" className="text-xs">Demo</Badge>
        </div>
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

      <section aria-label="Logs de auditoría" className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">Logs de auditoría</h2>
        </div>

        <div className="flex flex-wrap gap-3">
          <select
            value={actionFilter}
            onChange={(e) => { setActionFilter(e.target.value); setPage(1); }}
            className="rounded-xl border border-gray-200 bg-white px-4 py-2.5 text-sm font-medium text-gray-700 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
          >
            <option value="all">Todas las acciones</option>
            <option value="user_suspended">Usuario suspendido</option>
            <option value="user_reactivated">Usuario reactivado</option>
            <option value="user_role_updated">Rol actualizado</option>
            <option value="offering_deleted">Producto eliminado</option>
            <option value="report_created">Reporte creado</option>
            <option value="report_approved">Reporte aprobado</option>
            <option value="report_rejected">Reporte rechazado</option>
          </select>
          <select
            value={targetTypeFilter}
            onChange={(e) => { setTargetTypeFilter(e.target.value); setPage(1); }}
            className="rounded-xl border border-gray-200 bg-white px-4 py-2.5 text-sm font-medium text-gray-700 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
          >
            <option value="all">Todos los tipos</option>
            <option value="user">Usuario</option>
            <option value="offering">Producto</option>
            <option value="report">Reporte</option>
          </select>
        </div>

        <div className="rounded-xl border border-gray-100 bg-white overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm" role="table">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left font-semibold text-gray-500">Acción</th>
                  <th className="px-4 py-3 text-left font-semibold text-gray-500">Actor</th>
                  <th className="px-4 py-3 text-left font-semibold text-gray-500">Objetivo</th>
                  <th className="px-4 py-3 text-left font-semibold text-gray-500">Fecha</th>
                  <th className="px-4 py-3 text-left font-semibold text-gray-500">Detalles</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {auditLogs.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="px-4 py-12 text-center text-gray-500">
                      {actionFilter !== 'all' || targetTypeFilter !== 'all'
                        ? 'No hay logs que coincidan'
                        : 'No hay logs de auditoría'}
                    </td>
                  </tr>
                ) : (
                  auditLogs.map((log) => (
                    <tr key={log.id} className="hover:bg-gray-50">
                      <td className="px-4 py-3">
                        <Badge tone="blue">{ACTION_LABELS[log.action] || log.action}</Badge>
                      </td>
                      <td className="px-4 py-3 text-gray-500 font-mono text-xs">{log.actor_id}</td>
                      <td className="px-4 py-3">
                        <span className="font-mono text-xs text-gray-500">{log.target_id}</span>
                        <Badge tone="gray" className="ml-2 text-xs">
                          {TARGET_TYPE_LABELS[log.target_type] || log.target_type}
                        </Badge>
                      </td>
                      <td className="px-4 py-3 text-gray-500">
                        {new Date(log.created_at).toLocaleString('es-ES', { dateStyle: 'short', timeStyle: 'short' })}
                      </td>
                      <td className="px-4 py-3">
                        {log.metadata && (
                          <pre className="text-xs text-gray-400 bg-gray-50 p-2 rounded max-w-xs overflow-auto">
                            {JSON.stringify(log.metadata, null, 2)}
                          </pre>
                        )}
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
      </section>
    </div>
  )
}