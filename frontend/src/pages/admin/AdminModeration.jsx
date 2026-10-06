import { useState, useEffect, useCallback } from 'react'
import { reports } from '../../services/api'
import { cn } from '../../lib/cn'
import Icon from '../../components/ui/Icon'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'

const STATUS_MAP = {
  pending: 'Pendiente',
  approved: 'Aprobado',
  rejected: 'Rechazado',
}

const TARGET_TYPE_LABEL = {
  offering: 'Producto',
  user: 'Productor',
}

export default function AdminModeration() {
  const [reportItems, setReportItems] = useState([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize] = useState(20)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [statusFilter, setStatusFilter] = useState('all')
  const [targetTypeFilter, setTargetTypeFilter] = useState('all')

  const fetchReports = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const params = { page, page_size: pageSize }
      if (statusFilter !== 'all') params.status = statusFilter
      if (targetTypeFilter !== 'all') params.targetType = targetTypeFilter
      const data = await reports.list(params)
      setReportItems(data.items || [])
      setTotal(data.total || 0)
    } catch (e) {
      setError('No se pudieron cargar los reportes')
      console.error(e)
    } finally {
      setLoading(false)
    }
  }, [page, pageSize, statusFilter, targetTypeFilter])

  useEffect(() => {
    fetchReports()
  }, [fetchReports])

  const handleResolve = async (id, action) => {
    try {
      await reports.resolve(id, action)
      fetchReports()
    } catch (e) {
      alert('Error al resolver reporte')
      console.error(e)
    }
  }

const getSeverityTone = (severity) => severity === 'Alta' ? 'red' : severity === 'Media' ? 'amber' : 'gray'

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
        <Icon name="error" size={48} className="mx-auto text-red-500 mb-4" />
        <p className="text-gray-600">{error}</p>
        <Button onClick={fetchReports} className="mt-4">Reintentar</Button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900">Moderación</h1>
      </div>

      <div className="flex flex-wrap gap-3">
        <select
          value={statusFilter}
          onChange={(e) => { setStatusFilter(e.target.value); setPage(1); }}
          className="rounded-xl border border-gray-200 bg-white px-4 py-2.5 text-sm font-medium text-gray-700 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
        >
          <option value="all">Todos los estados</option>
          <option value="pending">Pendientes</option>
          <option value="approved">Aprobados</option>
          <option value="rejected">Rechazados</option>
        </select>
        <select
          value={targetTypeFilter}
          onChange={(e) => { setTargetTypeFilter(e.target.value); setPage(1); }}
          className="rounded-xl border border-gray-200 bg-white px-4 py-2.5 text-sm font-medium text-gray-700 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
        >
          <option value="all">Todos los tipos</option>
          <option value="offering">Productos</option>
          <option value="user">Productores</option>
        </select>
      </div>

      <div className="space-y-4">
        {reportItems.length === 0 ? (
          <div className="rounded-xl border border-gray-100 bg-white p-12 text-center text-gray-500">
            {statusFilter !== 'all' || targetTypeFilter !== 'all'
              ? 'No hay reportes que coincidan'
              : 'No hay reportes pendientes'}
          </div>
        ) : (
          reportItems.map((report) => (
            <article
              key={report.id}
              className={cn(
                'rounded-xl border border-gray-100 bg-white p-5',
                report.status !== 'pending' && 'opacity-50',
              )}
            >
              <div className="flex items-start justify-between gap-4">
                <div className="flex items-center gap-3 min-w-0 flex-1">
                  <span
                    className={cn(
                      'inline-flex h-2.5 w-2.5 rounded-full shrink-0',
                      report.severity === 'Alta' && 'bg-red-500',
                      report.severity === 'Media' && 'bg-yellow-500',
                      report.severity === 'Baja' && 'bg-gray-400',
                    )}
                  />
                  <div className="min-w-0">
                    <h3 className="font-semibold text-gray-900 truncate">{report.title}</h3>
                    <p className="text-sm text-gray-500">
                      {TARGET_TYPE_LABEL[report.target_type] || report.target_type} ·
                      Reportado por {report.reporter?.name || 'Desconocido'} ·
                      {new Date(report.created_at).toLocaleString('es-ES', { dateStyle: 'short', timeStyle: 'short' })}
                    </p>
                  </div>
                </div>
                <Badge tone={getSeverityTone(report.severity)}>
                  {report.severity}
                </Badge>
              </div>

              <p className="mt-3 text-sm text-gray-600 italic">"{report.reason}"</p>

              <div className="mt-4 flex items-center justify-end gap-2">
                {report.status === 'pending' ? (
                  <>
                    <Button variant="outline" size="sm" onClick={() => {}}>
                      Investigar
                    </Button>
                    <Button variant="outline" size="sm" className="text-red-600 hover:bg-red-50" onClick={() => handleResolve(report.id, 'reject')}>
                      Rechazar
                    </Button>
                    <Button variant="primary" size="sm" onClick={() => handleResolve(report.id, 'approve')}>
                      Aprobar
                    </Button>
                  </>
                ) : (
                  <span className="inline-flex items-center gap-1.5 text-sm font-medium text-green-600">
                    <Icon name="check_circle" size={16} />
                    {STATUS_MAP[report.status] || 'Resuelto'}
                  </span>
                )}
              </div>
            </article>
          ))
        )}
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
  )
}