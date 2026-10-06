import { useState, useEffect } from 'react'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'
import Icon from '../../components/ui/Icon'
import MessagesInbox from '../../components/messages/MessagesInbox'
import { inquiries } from '../../services/api'
import { useAuth } from '../../context/AuthContext'
import { cn } from '../../lib/cn'

const STATUS_MAP = {
  pending: { label: 'Pendiente', tone: 'amber' },
  read: { label: 'Leída', tone: 'brand' },
  replied: { label: 'Respondida', tone: 'brand' },
  closed: { label: 'Cerrada', tone: 'gray' },
}

const TABS = [
  { id: 'conversations', label: 'Conversaciones' },
  { id: 'inquiries', label: 'Consultas' },
]

export default function BuyerMessages() {
  const { user } = useAuth()
  const [tab, setTab] = useState('conversations')
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  function fetchInquiries() {
    setLoading(true)
    setError('')
    inquiries.getByUser(user?.id)
      .then((data) => {
        setItems(Array.isArray(data) ? data : [])
      })
      .catch((err) => {
        setError(err.message || 'Error al cargar mensajes.')
      })
      .finally(() => {
        setLoading(false)
      })
  }

  useEffect(() => {
    if (tab !== 'inquiries') return
    fetchInquiries()
  }, [tab, user?.id])

  function handleClose(id) {
    inquiries.updateStatus(id, 'closed')
      .then(() => fetchInquiries())
      .catch((err) => alert(err.message || 'Error al actualizar.'))
  }

  return (
    <div className="space-y-6">
      <header>
        <h1 className="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">Mensajes</h1>
        <p className="mt-1 text-sm text-gray-500">
          Conversaciones con agricultores y consultas que enviaste.
        </p>
      </header>

      <div className="flex gap-2 border-b border-gray-100 pb-3" aria-label="Secciones de mensajes">
        {TABS.map((item) => (
          <button
            key={item.id}
            type="button"
            aria-pressed={tab === item.id}
            onClick={() => setTab(item.id)}
            className={cn(
              'rounded-full px-4 py-1.5 text-sm font-semibold transition-colors',
              tab === item.id
                ? 'bg-brand text-white'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200',
            )}
          >
            {item.label}
          </button>
        ))}
      </div>

      {tab === 'conversations' ? (
        <MessagesInbox />
      ) : (
        <InquiriesPane
          items={items}
          loading={loading}
          error={error}
          onRetry={fetchInquiries}
          onClose={handleClose}
        />
      )}
    </div>
  )
}

function InquiriesPane({ items, loading, error, onRetry, onClose }) {
  if (loading) {
    return (
      <div className="space-y-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="animate-pulse rounded-xl border border-gray-100 bg-white p-5">
            <div className="flex items-start justify-between gap-4">
              <div className="flex-1 space-y-2">
                <div className="h-4 w-40 rounded bg-gray-200" />
                <div className="h-3 w-64 rounded bg-gray-200" />
              </div>
              <div className="h-6 w-20 rounded bg-gray-200" />
            </div>
            <div className="mt-3 h-3 w-full rounded bg-gray-200" />
          </div>
        ))}
      </div>
    )
  }

  if (error) {
    return (
      <div className="rounded-xl border border-red-200 bg-red-50 p-6 text-center">
        <Icon name="error" size={40} className="mx-auto text-red-400" />
        <p className="mt-3 text-sm text-red-700">{error}</p>
        <Button variant="outline" size="sm" onClick={onRetry} className="mt-4">
          Reintentar
        </Button>
      </div>
    )
  }

  if (items.length === 0) {
    return (
      <div className="rounded-xl border border-gray-100 bg-white p-12 text-center">
        <Icon name="chat_bubble" size={48} className="mx-auto text-gray-300" />
        <h2 className="mt-4 text-lg font-semibold text-gray-900">Sin mensajes</h2>
        <p className="mt-2 text-sm text-gray-500">
          Cuando envíes una consulta desde un producto, aparecerá aquí.
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {items.map((item) => {
        const status = STATUS_MAP[item.status] || STATUS_MAP.pending
        return (
          <article
            key={item.id}
            className="rounded-xl border border-gray-100 bg-white p-5 transition-colors hover:shadow-sm"
          >
            <div className="flex items-start justify-between gap-4">
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                  <a
                    href={`#/product/${item.offering_id}`}
                    className="font-semibold text-gray-900 hover:text-brand hover:underline"
                  >
                    {item.offering_name || 'Producto'}
                  </a>
                </div>
                <time className="mt-0.5 block text-xs text-gray-400">
                  {new Date(item.created_at).toLocaleDateString('es-NI', {
                    day: 'numeric',
                    month: 'short',
                    year: 'numeric',
                  })}
                </time>
              </div>

              <Badge tone={status.tone} className="shrink-0">
                {status.label}
              </Badge>
            </div>

            <p className="mt-3 text-sm text-gray-600 italic leading-relaxed">"{item.message}"</p>

            {item.status === 'replied' && (
              <div className="mt-3 rounded-lg border border-green-200 bg-green-50 p-3">
                <p className="text-sm text-green-700">
                  El agricultor ha respondido. Para ver la respuesta, contacta directamente por WhatsApp desde la página del producto.
                </p>
                <a
                  href={`#/product/${item.offering_id}`}
                  className="mt-2 inline-flex items-center gap-1 text-sm font-semibold text-green-800 hover:underline"
                >
                  <Icon name="open_in_new" size={14} />
                  Ver producto
                </a>
              </div>
            )}

            {item.status === 'pending' && (
              <div className="mt-4 flex items-center justify-end gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => onClose(item.id)}
                >
                  Cerrar consulta
                </Button>
              </div>
            )}
          </article>
        )
      })}
    </div>
  )
}
