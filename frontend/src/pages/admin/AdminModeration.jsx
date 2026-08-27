import { useState } from 'react'
import { cn } from '../../lib/cn'
import Icon from '../../components/ui/Icon'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'
import { adminModerationReports } from '../../mocks/admin'

export default function AdminModeration() {
  const [reports] = useState(adminModerationReports.map((r) => ({ ...r })))

  function toggleStatus(reportId) {
    const report = reports.find((r) => r.id === reportId)
    if (report) {
      report.status = report.status === 'pending' ? 'resolved' : 'pending'
    }
  }

  function getSeverityTone(severity) {
    return severity === 'Alta' ? 'red' : severity === 'Media' ? 'amber' : 'gray'
  }

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold text-gray-900">Moderación</h1>

      <div className="space-y-4">
        {reports.map((report) => (
          <article
            key={report.id}
            className={cn(
              'rounded-xl border border-gray-100 bg-white p-5',
              report.status === 'resolved' && 'opacity-50',
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
                  <p className="text-sm text-gray-500">{report.type} · Reportado por {report.reportedBy} · {report.time}</p>
                </div>
              </div>
              <Badge tone={getSeverityTone(report.severity)}>
                {report.severity}
              </Badge>
            </div>

            <p className="mt-3 text-sm text-gray-600 italic">"{report.detail}"</p>

            <div className="mt-4 flex items-center justify-end gap-2">
              {report.status === 'pending' ? (
                <>
                  <Button variant="outline" size="sm" onClick={() => toggleStatus(report.id)}>
                    Investigar
                  </Button>
                  <Button variant="outline" size="sm" className="text-red-600 hover:bg-red-50">
                    Eliminar contenido
                  </Button>
                  <Button variant="primary" size="sm" onClick={() => toggleStatus(report.id)}>
                    Resolver
                  </Button>
                </>
              ) : (
                <span className="inline-flex items-center gap-1.5 text-sm font-medium text-green-600">
                  <Icon name="check_circle" size={16} />
                  Reporte resuelto
                </span>
              )}
            </div>
          </article>
        ))}
      </div>
    </div>
  )
}