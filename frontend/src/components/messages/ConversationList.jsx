import { useCallback, useEffect, useState } from 'react'
import Avatar from '../Avatar'
import Badge from '../ui/Badge'
import Button from '../ui/Button'
import EmptyState from '../ui/EmptyState'
import Icon from '../ui/Icon'
import Skeleton from '../ui/Skeleton'
import { cn } from '../../lib/cn'
import { conversations, offerings, users } from '../../services/api'
import { useAuth } from '../../context/AuthContext'

const NIL_UUID = '00000000-0000-0000-0000-000000000000'

function peerIdOf(conversation, myId) {
  return conversation.buyer_id === myId ? conversation.farmer_id : conversation.buyer_id
}

function nameOf(user) {
  const full = [user?.first_name, user?.last_name].filter(Boolean).join(' ').trim()
  return full || 'Participante'
}

function initialsOf(name) {
  const parts = String(name || '')
    .split(/\s+/)
    .filter(Boolean)
  if (!parts.length) return '?'
  return (parts[0][0] + (parts[1]?.[0] || '')).toUpperCase()
}

function formatDate(value) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleDateString('es-NI', { day: 'numeric', month: 'short' })
}

export default function ConversationList({ onSelect, selectedId, className }) {
  const { user } = useAuth()
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [attempt, setAttempt] = useState(0)

  const reload = useCallback(() => setAttempt((n) => n + 1), [])

  useEffect(() => {
    let alive = true
    setLoading(true)
    setError('')

    conversations
      .list()
      .then(async (list) => {
        const rows = Array.isArray(list) ? list : []
        const enriched = await Promise.all(
          rows.map(async (conversation) => {
            const peerId = peerIdOf(conversation, user?.id)
            const offeringId =
              conversation.offering_id && conversation.offering_id !== NIL_UUID
                ? conversation.offering_id
                : null
            const [peer, offering] = await Promise.all([
              users.getById(peerId).catch(() => null),
              offeringId ? offerings.getById(offeringId).catch(() => null) : null,
            ])
            return {
              ...conversation,
              peerId,
              peerName: nameOf(peer),
              offeringName: offering?.name || '',
            }
          }),
        )
        enriched.sort(
          (a, b) => new Date(b.updated_at || b.created_at) - new Date(a.updated_at || a.created_at),
        )
        if (!alive) return
        setItems(enriched)
        setLoading(false)
      })
      .catch((err) => {
        if (!alive) return
        setError(err?.message || 'No se pudieron cargar las conversaciones.')
        setLoading(false)
      })

    return () => {
      alive = false
    }
  }, [attempt, user?.id])

  if (loading) {
    return (
      <div className={cn('space-y-2', className)}>
        {[1, 2, 3].map((i) => (
          <div key={i} className="flex items-center gap-3 rounded-xl border border-gray-100 bg-white p-4">
            <Skeleton className="h-10 w-10 rounded-full" />
            <div className="flex-1 space-y-2">
              <Skeleton className="h-4 w-32" />
              <Skeleton className="h-3 w-40" />
            </div>
          </div>
        ))}
      </div>
    )
  }

  if (error) {
    return (
      <div className={cn('rounded-xl border border-red-200 bg-red-50 p-6 text-center', className)}>
        <Icon name="error" size={32} className="mx-auto text-red-400" />
        <p className="mt-3 text-sm text-red-700">{error}</p>
        <Button variant="outline" size="sm" onClick={reload} className="mt-4">
          Reintentar
        </Button>
      </div>
    )
  }

  if (!items.length) {
    return (
      <EmptyState
        icon="chat"
        title="Sin conversaciones"
        description="Cuando alguien haga match o te contactes desde un producto, la conversación aparecerá aquí."
        className={className}
      />
    )
  }

  return (
    <div className={cn('space-y-2', className)} aria-label="Lista de conversaciones">
      {items.map((conversation) => {
        const selected = conversation.id === selectedId
        const isMatch = Boolean(conversation.match_id)
        return (
          <button
            key={conversation.id}
            type="button"
            aria-pressed={selected}
            onClick={() => onSelect?.(conversation)}
            className={cn(
              'flex w-full items-center gap-3 rounded-xl border p-4 text-left transition-colors',
              selected
                ? 'border-brand bg-brand-soft/60'
                : 'border-gray-100 bg-white hover:border-brand/40 hover:shadow-sm',
            )}
          >
            <Avatar initials={initialsOf(conversation.peerName)} name={conversation.peerName} size="md" />
            <span className="min-w-0 flex-1">
              <span className="flex items-center justify-between gap-2">
                <span className="truncate font-semibold text-gray-900">{conversation.peerName}</span>
                <time className="shrink-0 text-xs text-gray-400">
                  {formatDate(conversation.updated_at || conversation.created_at)}
                </time>
              </span>
              <span className="mt-0.5 flex items-center gap-2">
                <Badge tone={isMatch ? 'brand' : 'gray'} className="shrink-0">
                  {isMatch ? 'Match' : conversation.offeringName || 'Producto'}
                </Badge>
                {!isMatch && conversation.offeringName && (
                  <span className="truncate text-xs text-gray-500">Consulta de producto</span>
                )}
              </span>
            </span>
          </button>
        )
      })}
    </div>
  )
}
