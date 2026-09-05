import { useState, useEffect } from 'react'
import { cn } from '../../lib/cn'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'
import Avatar from '../../components/Avatar'
import Icon from '../../components/ui/Icon'
import { inquiries } from '../../services/api'

const STATUS_MAP = {
  pending: { label: 'Pendiente', tone: 'amber' },
  read: { label: 'Leída', tone: 'brand' },
  replied: { label: 'Respondida', tone: 'brand' },
  closed: { label: 'Cerrada', tone: 'gray' },
}

export default function ProducerRequests() {
  const [requests, setRequests] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  function fetchRequests() {
    setLoading(true)
    setError('')
    inquiries.getByProducer()
      .then((data) => {
        setRequests(Array.isArray(data) ? data : data.inquiries || [])
      })
      .catch((err) => {
        setError(err.message || 'Error al cargar solicitudes.')
      })
      .finally(() => {
        setLoading(false)
      })
  }

  useEffect(() => {
    fetchRequests()
  }, [])

  function handleReply(id) {
    inquiries.updateStatus(id, 'replied')
      .then(() => {
        fetchRequests()
      })
      .catch((err) => {
        alert(err.message || 'Error al actualizar.')
      })
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <h1 className="text-3xl font-bold text-gray-900">Solicitudes de compra</h1>
        <div className="space-y-4">
          {[1, 2, 3].map((i) => (
            <div key={i} className="animate-pulse rounded-xl border border-gray-100 bg-white p-5">
              <div className="flex items-start justify-between gap-4">
                <div className="flex items-center gap-3 flex-1">
                  <div className="h-12 w-12 rounded-full bg-gray-200" />
                  <div className="flex-1 space-y-2">
                    <div className="h-4 w-32 rounded bg-gray-200" />
                    <div className="h-3 w-48 rounded bg-gray-200" />
                  </div>
                </div>
                <div className="h-6 w-20 rounded bg-gray-200" />
              </div>
              <div className="mt-4 h-3 w-full rounded bg-gray-200" />
            </div>
          ))}
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-6">
        <h1 className="text-3xl font-bold text-gray-900">Solicitudes de compra</h1>
        <div className="rounded-xl border border-red-200 bg-red-50 p-6 text-center">
          <Icon name="error" size={40} className="mx-auto text-red-400" />
          <p className="mt-3 text-sm text-red-700">{error}</p>
          <Button variant="outline" size="sm" onClick={fetchRequests} className="mt-4">
            Reintentar
          </Button>
        </div>
      </div>
    )
  }

  if (requests.length === 0) {
    return (
      <div className="space-y-6">
        <h1 className="text-3xl font-bold text-gray-900">Solicitudes de compra</h1>
        <div className="rounded-xl border border-gray-100 bg-white p-12 text-center">
          <Icon name="inbox" size={48} className="mx-auto text-gray-300" />
          <h2 className="mt-4 text-lg font-semibold text-gray-900">Sin solicitudes</h2>
          <p className="mt-2 text-sm text-gray-500">
            Cuando los compradores contacten por tus productos, las solicitudes aparecerán aquí.
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold text-gray-900">Solicitudes de compra</h1>

      <div className="space-y-4">
        {requests.map((req) => {
          const status = STATUS_MAP[req.status] || STATUS_MAP.pending
          return (
            <article
              key={req.id}
              className={cn(
                'rounded-xl border border-gray-100 bg-white p-5 transition-colors',
                req.status === 'pending' ? 'border-amber-200' : 'border-gray-100',
              )}
            >
              <div className="flex items-start justify-between gap-4">
                <div className="flex items-center gap-3 min-w-0 flex-1">
                  <Avatar
                    initials={(req.buyer?.first_name?.[0] || '') + (req.buyer?.last_name?.[0] || '')}
                    name={`${req.buyer?.first_name || ''} ${req.buyer?.last_name || ''}`}
                    size="md"
                    tone="brand"
                  />
                  <div className="min-w-0">
                    <p className="font-semibold text-gray-900 truncate">
                      {req.buyer?.first_name} {req.buyer?.last_name}
                    </p>
                    <a href={`#/product/${req.offering_id}`} className="text-sm font-medium text-brand hover:underline truncate block">
                      Producto #{req.offering_id}
                    </a>
                  </div>
                </div>

                <div className="flex items-center gap-3 shrink-0">
                  <Badge tone={status.tone} className="shrink-0">
                    {status.label}
                  </Badge>
                  <span className="text-sm text-gray-500 whitespace-nowrap shrink-0">
                    {new Date(req.created_at).toLocaleDateString('es-NI')}
                  </span>
                </div>
              </div>

              <p className="mt-3 text-sm text-gray-600 italic">"{req.message}"</p>

              {req.status === 'pending' && (
                <div className="mt-4 flex items-center justify-end gap-2">
                  <Button variant="outline" size="sm" onClick={() => window.location.hash = `#/product/${req.offering_id}`}>
                    Ver producto
                  </Button>
                  <Button variant="primary" size="sm" onClick={() => handleReply(req.id)}>
                    Responder
                  </Button>
                </div>
              )}
            </article>
          )
        })}
      </div>
    </div>
  )
}
