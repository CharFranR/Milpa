import { useEffect, useRef, useState } from 'react'
import { useAuth } from '../../context/AuthContext'
import { useChat } from '../../hooks/useChat'
import { cn } from '../../lib/cn'
import Button from '../ui/Button'
import Icon from '../ui/Icon'
import Skeleton from '../ui/Skeleton'

const STATUS = {
  idle: { label: 'Inactivo', tone: 'bg-gray-200' },
  loading: { label: 'Cargando…', tone: 'bg-amber-300' },
  connecting: { label: 'Conectando…', tone: 'bg-amber-400' },
  connected: { label: 'Conectado', tone: 'bg-green-500' },
  reconnecting: { label: 'Reconectando…', tone: 'bg-amber-400' },
  error: { label: 'Sin conexión', tone: 'bg-red-500' },
}

export default function ChatPanel({
  conversationId,
  title = 'Conversación',
  hint = 'Escribe tu mensaje. El otro participante responderá por aquí.',
  onClose,
  className,
}) {
  const { user } = useAuth()
  const { messages, status, error, send, reloadHistory } = useChat(conversationId)
  const [draft, setDraft] = useState('')
  const listRef = useRef(null)

  useEffect(() => {
    const node = listRef.current
    if (node) node.scrollTop = node.scrollHeight
  }, [messages])

  function handleSubmit(event) {
    event.preventDefault()
    const text = draft.trim()
    if (!text) return
    if (send(text)) setDraft('')
  }

  if (!conversationId) return null

  const statusInfo = STATUS[status] || STATUS.idle
  const ready = status === 'connected'

  return (
    <div className={cn('flex h-full min-h-72 flex-col rounded-xl border border-gray-100 bg-white', className)}>
      <header className="flex items-center justify-between gap-3 border-b border-gray-100 px-4 py-3">
        <h3 className="truncate text-sm font-semibold text-gray-900">{title}</h3>
        <span className="flex shrink-0 items-center gap-3 text-xs text-gray-500">
          <span className="flex items-center gap-2">
            <span aria-hidden className={cn('h-2 w-2 rounded-full', statusInfo.tone)} />
            {statusInfo.label}
          </span>
          {onClose && (
            <button
              type="button"
              onClick={onClose}
              aria-label="Cerrar"
              className="rounded-lg p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
            >
              <Icon name="close" size={18} />
            </button>
          )}
        </span>
      </header>

      <div ref={listRef} className="max-h-80 flex-1 space-y-2 overflow-y-auto p-4" aria-live="polite">
        {status === 'loading' && (
          <div className="space-y-2">
            <Skeleton className="h-10 w-2/3" />
            <Skeleton className="ml-auto h-10 w-1/2" />
            <Skeleton className="h-10 w-3/5" />
          </div>
        )}

        {status !== 'loading' && messages.length === 0 && (
          <p className="text-sm text-gray-500">{hint}</p>
        )}

        {messages.map((message, index) => {
          const mine = message.sender_id === user?.id
          return (
            <div
              key={message.id || `msg-${index}`}
              className={cn('flex', mine ? 'justify-end' : 'justify-start')}
            >
              <p
                className={cn(
                  'max-w-[80%] whitespace-pre-wrap rounded-2xl px-3 py-2 text-sm',
                  mine ? 'bg-brand text-white' : 'bg-gray-50 text-gray-800',
                )}
              >
                {message.content}
              </p>
            </div>
          )
        })}
      </div>

      {error && (
        <div className="flex items-center justify-between gap-3 border-t border-gray-100 px-4 py-2">
          <p role="alert" className="text-xs text-red-600">
            {error}
          </p>
          <button
            type="button"
            onClick={reloadHistory}
            className="text-xs font-semibold text-brand hover:underline"
          >
            Reintentar
          </button>
        </div>
      )}

      <form onSubmit={handleSubmit} className="flex items-center gap-2 border-t border-gray-100 p-3">
        <label htmlFor={`chat-message-${conversationId}`} className="sr-only">
          Tu mensaje
        </label>
        <input
          id={`chat-message-${conversationId}`}
          value={draft}
          onChange={(event) => setDraft(event.target.value)}
          placeholder="Escribe un mensaje..."
          className="flex-1 rounded-xl border border-gray-200 px-4 py-2 text-sm text-gray-900 placeholder:text-gray-400 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
          disabled={!ready}
          maxLength={2000}
        />
        <Button type="submit" variant="primary" disabled={!ready || !draft.trim()}>
          Enviar
        </Button>
      </form>

      {!ready && status !== 'loading' && (
        <p className="px-4 pb-3 text-xs text-gray-400">
          <Icon name="info" size={13} /> Se enviarán mensajes cuando la conexión vuelva.
        </p>
      )}
    </div>
  )
}
