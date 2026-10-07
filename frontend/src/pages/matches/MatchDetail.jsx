import { useCallback, useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import ChatPanel from '../../components/chat/ChatPanel'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'
import ConfirmDialog from '../../components/ui/ConfirmDialog'
import EmptyState from '../../components/ui/EmptyState'
import ErrorState from '../../components/ui/ErrorState'
import Icon from '../../components/ui/Icon'
import Skeleton from '../../components/ui/Skeleton'
import Spinner from '../../components/ui/Spinner'
import Toast from '../../components/ui/Toast'
import ReviewForm from '../../components/reviews/ReviewForm'
import ReportForm from '../../components/reports/ReportForm'
import StarRating from '../../components/StarRating'
import { useAuth } from '../../context/AuthContext'
import { formatPrice } from '../../lib/format'
import { friendlyTransactionError } from '../../lib/transactionMessages'
import {
  formatDateTime,
  isZeroTime,
  matchStatus,
  measurementLabel,
  transactionStatus,
} from '../../lib/supplyStatus'
import { conversations } from '../../services/conversations'
import { companies } from '../../services/companies'
import { matches } from '../../services/matches'
import { reviews } from '../../services/reviews'
import { reports } from '../../services/reports'
import { supplyOffers } from '../../services/supplyOffers'
import { supplyRequests } from '../../services/supplyRequests'
import { transactions } from '../../services/transactions'
import { isBuyer } from '../../lib/roles'

const TIMELINE_LABELS = {
  0: 'Match creado',
  1: 'En progreso',
  2: 'Completada',
  3: 'Cancelada',
}

// El servidor cachea la lista de conversaciones 5 minutos y el match se crea
// después del último paso de caché, así que la conversación se busca con
// reintentos que cubren esa ventana antes de rendirse.
const CHAT_RETRY_DELAYS = [0, 10, 30, 60, 120, 180, 240, 300]

function ConfirmRow({ title, buyerAt, supplierAt }) {
  const buyerDone = !isZeroTime(buyerAt)
  const supplierDone = !isZeroTime(supplierAt)
  return (
    <div className="flex flex-wrap items-center justify-between gap-2 rounded-lg bg-gray-50 px-4 py-2.5">
      <span className="text-sm font-medium text-gray-700">{title}</span>
      <span className="flex items-center gap-3 text-xs text-gray-500">
        <span className={buyerDone ? 'text-green-600' : 'text-gray-400'}>
          Comprador: {buyerDone ? formatDateTime(buyerAt) : 'pendiente'}
        </span>
        <span className={supplierDone ? 'text-green-600' : 'text-gray-400'}>
          Agricultor: {supplierDone ? formatDateTime(supplierAt) : 'pendiente'}
        </span>
      </span>
    </div>
  )
}

export default function MatchDetail() {
  const { matchId } = useParams()
  const { role, user } = useAuth()
  const buyerRole = isBuyer(role)

  const [match, setMatch] = useState(null)
  const [transaction, setTransaction] = useState(null)
  const [conversationId, setConversationId] = useState(null)
  const [chatState, setChatState] = useState('searching')
  const [chatAttempt, setChatAttempt] = useState(0)
  const [myReviews, setMyReviews] = useState([])
  const [counterpartyId, setCounterpartyId] = useState(null)
  const [counterpartyCompany, setCounterpartyCompany] = useState(null)
  const [reporting, setReporting] = useState(false)
  const [reported, setReported] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)
  const [cancelOpen, setCancelOpen] = useState(false)
  const [reason, setReason] = useState('')
  const [reasonError, setReasonError] = useState('')
  const [toast, setToast] = useState(null)

  const reload = useCallback(() => {
    if (!matchId) return
    setLoading(true)
    setError('')
    Promise.all([
      matches.getById(matchId),
      transactions.getByMatch(matchId),
      user?.id ? reviews.listByUser(user.id).catch(() => []) : Promise.resolve([]),
    ])
      .then(([matchData, transactionData, reviewList]) => {
        setMatch(matchData)
        setTransaction(transactionData)
        setMyReviews(Array.isArray(reviewList) ? reviewList : [])
      })
      .catch((err) => setError(err.message || 'No se pudo cargar el match.'))
      .finally(() => setLoading(false))
  }, [matchId, user?.id])

  useEffect(() => {
    reload()
  }, [reload])

  useEffect(() => {
    if (!matchId || !match || conversationId) return undefined
    let alive = true
    let timer = null

    const search = (step) => {
      conversations.list().then(
        (list) => {
          if (!alive) return
          const rows = Array.isArray(list) ? list : []
          const found = rows.find((item) => item.match_id === matchId)
          if (found) {
            setConversationId(found.id)
            return
          }
          const next = step + 1
          if (next >= CHAT_RETRY_DELAYS.length) {
            setChatState('missing')
            return
          }
          setChatState('searching')
          timer = setTimeout(() => search(next), CHAT_RETRY_DELAYS[next] * 1000)
        },
        () => {
          if (!alive) return
          const next = Math.min(step + 1, CHAT_RETRY_DELAYS.length - 1)
          setChatState('searching')
          timer = setTimeout(() => search(next), CHAT_RETRY_DELAYS[next] * 1000)
        },
      )
    }

    setChatState('searching')
    search(0)

    return () => {
      alive = false
      if (timer) clearTimeout(timer)
    }
  }, [match, matchId, conversationId, chatAttempt])

  useEffect(() => {
    if (!match) return undefined
    let cancelled = false

    Promise.all([
      supplyRequests.getById(match.supply_request).catch(() => null),
      supplyOffers.getById(match.supply_offer).catch(() => null),
    ]).then(([requestData, offerData]) => {
      if (cancelled) return
      const myId = user?.id
      let other = buyerRole ? offerData?.supplier_id : requestData?.buyer_id
      if (other && other === myId) {
        other = buyerRole ? requestData?.buyer_id : offerData?.supplier_id
      }
      setCounterpartyId(other || null)
      setCounterpartyCompany(null)
    })

    return () => {
      cancelled = true
    }
  }, [match, buyerRole, user?.id])

  useEffect(() => {
    if (!counterpartyId || transaction?.status !== 2) return undefined
    let cancelled = false

    companies.getByOwner(counterpartyId)
      .then((list) => {
        if (cancelled) return
        setCounterpartyCompany(Array.isArray(list) && list.length > 0 ? list[0] : null)
      })
      .catch(() => {
        if (cancelled) return
        setCounterpartyCompany(null)
      })

    return () => {
      cancelled = true
    }
  }, [counterpartyId, transaction?.status])

  async function run(action) {
    if (!transaction) return
    setBusy(true)
    try {
      if (action === 'start') await transactions.confirmStart(transaction.id)
      else if (action === 'delivery') await transactions.confirmDelivery(transaction.id)
      else await transactions.cancel(transaction.id, reason.trim())

      setCancelOpen(false)
      setReason('')
      setReasonError('')
      setToast({
        message:
          action === 'start'
            ? 'Inicio confirmado.'
            : action === 'delivery'
              ? 'Entrega confirmada.'
              : 'Transacción cancelada.',
        tone: 'success',
      })
      reload()
    } catch (err) {
      const message = friendlyTransactionError(err)
      if (action === 'cancel') setReasonError(message)
      else setToast({ message, tone: 'error' })
      setCancelOpen(action === 'cancel')
    } finally {
      setBusy(false)
    }
  }

  function openCancel() {
    setReason('')
    setReasonError('')
    setCancelOpen(true)
  }

  async function submitReview({ rating, comment }) {
    if (!transaction || !reviewTarget) {
      throw new Error('Faltan datos para guardar la reseña.')
    }

    await reviews.create({
      transaction_id: transaction.id,
      target_type: reviewTarget.type,
      target_id: reviewTarget.id,
      rating,
      comment,
    })

    setToast({ message: 'Reseña enviada.', tone: 'success' })
    reload()
  }

  async function submitUserReport({ reason }) {
    if (!counterpartyId) {
      throw new Error('Faltan datos para enviar la denuncia.')
    }

    await reports.create({
      target_type: 'user',
      target_id: counterpartyId,
      reason,
    })

    setReporting(false)
    setReported(true)
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-6 w-40" />
        <div className="rounded-2xl border border-gray-100 bg-white p-6">
          <Skeleton className="h-5 w-64" />
          <Skeleton className="mt-3 h-4 w-48" />
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-6">
        <ErrorState message={error} onRetry={reload} />
        <Link
          to={buyerRole ? '/dashboard' : '/producer/offers'}
          className="inline-flex items-center gap-1 text-sm font-semibold text-brand hover:underline"
        >
          <Icon name="chevron_left" size={16} />
          Volver
        </Link>
      </div>
    )
  }

  if (!match) return null

  const mStatus = matchStatus(match.status)
  const tStatus = transaction ? transactionStatus(transaction.status) : null
  const tx = transaction
  const status = tx?.status

  const startMine = buyerRole ? tx?.buyer_start_confirmed_at : tx?.supplier_start_confirmed_at
  const deliveryMine = buyerRole ? tx?.buyer_delivery_confirmed_at : tx?.supplier_delivery_confirmed_at
  const canConfirmStart = status === 0 && isZeroTime(startMine)
  const canConfirmDelivery = status === 1 && isZeroTime(deliveryMine)
  const canCancel = status === 0 || status === 1
  const myReview = tx ? myReviews.find((item) => item.transaction_id === tx.id) : null
  const reviewTarget = counterpartyCompany
    ? { type: 'company', id: counterpartyCompany.id }
    : counterpartyId
      ? { type: 'user', id: counterpartyId }
      : null
  const alreadyReviewedTarget = reviewTarget
    ? myReviews.some((item) => item.target_type === reviewTarget.type && item.target_id === reviewTarget.id)
    : false
  const reviewTargetLabel = counterpartyCompany
    ? `a la empresa ${counterpartyCompany.name}`
    : buyerRole
      ? 'al agricultor'
      : 'al comprador'

  const backTo = buyerRole
    ? `/dashboard/requests/${match.supply_request}`
    : '/producer/offers'

  return (
    <div className="space-y-6">
      <header>
        <Link
          to={backTo}
          className="inline-flex items-center gap-1 text-sm font-semibold text-brand hover:underline"
        >
          <Icon name="chevron_left" size={16} />
          {buyerRole ? 'Volver a la solicitud' : 'Volver a mis ofertas'}
        </Link>
        <div className="mt-2 flex flex-wrap items-center gap-3">
          <h1 className="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">
            Match confirmado
          </h1>
          <Badge tone={mStatus.tone}>{mStatus.label}</Badge>
        </div>
        <p className="mt-1 text-sm text-gray-500">
          Creado el {formatDateTime(match.created_at)} · {String(match.id).slice(0, 8)}
        </p>
      </header>

      <section className="rounded-2xl border border-gray-100 bg-white p-6">
        <h2 className="text-base font-bold text-gray-900">Monto acordado</h2>
        <p className="mt-2 text-xl font-semibold text-gray-900">
          {formatPrice(match.matched_amount)} {measurementLabel(match.amount_unit)}
        </p>
        <p className="mt-1 text-sm text-gray-500">
          Reservado de la solicitud. Confirma el inicio cuando el agricultor despache.
        </p>
      </section>

      <section aria-label="Conversación" className="space-y-3">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <h2 className="text-base font-bold text-gray-900">Conversación</h2>
          {conversationId && counterpartyId && !reported && (
            <button
              type="button"
              onClick={() => setReporting((value) => !value)}
              className="inline-flex items-center gap-1 text-sm font-semibold text-red-600 hover:text-red-700"
            >
              <Icon name="flag" size={16} />
              Reportar {buyerRole ? 'al agricultor' : 'al comprador'}
            </button>
          )}
        </div>

        {reported && (
          <p className="rounded-lg bg-amber-50 px-3 py-2 text-xs text-amber-800">
            Denuncia enviada. El equipo de moderación revisará este caso.
          </p>
        )}

        {reporting && !reported && (
          <div className="rounded-xl border border-gray-100 bg-white p-4">
            <ReportForm
              targetLabel={buyerRole ? 'al agricultor' : 'al comprador'}
              onSubmit={submitUserReport}
              onCancel={() => setReporting(false)}
            />
          </div>
        )}

        {conversationId ? (
          <ChatPanel
            conversationId={conversationId}
            title={buyerRole ? 'Con el agricultor' : 'Con el comprador'}
            hint="La conversación se creó con el match. Coordina aquí el despacho."
          />
        ) : chatState === 'missing' ? (
          <EmptyState
            icon="chat"
            title="Chat no disponible todavía"
            description="La conversación del match todavía no aparece en tu bandeja. El servidor la guarda en caché unos minutos; reintenta en un momento."
            action={
              <Button variant="outline" onClick={() => setChatAttempt((value) => value + 1)}>
                Reintentar ahora
              </Button>
            }
          />
        ) : (
          <div className="flex flex-col items-center gap-3 rounded-2xl border border-gray-100 bg-white p-10 text-center">
            <Spinner size={28} label="Buscando el chat del match…" />
            <p className="max-w-md text-sm text-gray-500">
              Estamos comprobando si la conversación ya está en tu bandeja.
            </p>
            <Button variant="outline" size="sm" onClick={() => setChatAttempt((value) => value + 1)}>
              Buscar ahora
            </Button>
          </div>
        )}
      </section>

      {tx && (
        <section className="rounded-2xl border border-gray-100 bg-white p-6 space-y-4">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <h2 className="text-base font-bold text-gray-900">Transacción</h2>
            <Badge tone={tStatus.tone}>{tStatus.label}</Badge>
          </div>

          <div className="space-y-2">
            <ConfirmRow
              title="Confirmación de inicio"
              buyerAt={tx.buyer_start_confirmed_at}
              supplierAt={tx.supplier_start_confirmed_at}
            />
            <ConfirmRow
              title="Confirmación de entrega"
              buyerAt={tx.buyer_delivery_confirmed_at}
              supplierAt={tx.supplier_delivery_confirmed_at}
            />
          </div>

          {status === 3 && tx.cancel_reason && (
            <p className="rounded-lg bg-red-50 px-4 py-3 text-sm text-red-700">
              Cancelada: {tx.cancel_reason}
            </p>
          )}

          <div className="flex flex-wrap gap-3">
            <Button onClick={() => run('start')} disabled={busy || !canConfirmStart}>
              {busy ? 'Confirmando…' : 'Confirmar inicio'}
            </Button>
            <Button onClick={() => run('delivery')} disabled={busy || !canConfirmDelivery}>
              Confirmar entrega
            </Button>
            <Button variant="danger" onClick={openCancel} disabled={busy || !canCancel}>
              Cancelar transacción
            </Button>
          </div>

          {Array.isArray(tx.history) && tx.history.length > 0 && (
            <div>
              <h3 className="text-xs font-semibold uppercase tracking-wide text-gray-400">
                Historial
              </h3>
              <ol className="mt-3 space-y-3 border-l-2 border-gray-100 pl-4">
                {tx.history.map((entry, index) => {
                  const entryStatus = transactionStatus(entry.status)
                  return (
                    <li key={`${entry.status}-${entry.at}-${index}`} className="relative">
                      <span
                        aria-hidden
                        className="absolute -left-[21px] top-1.5 h-2.5 w-2.5 rounded-full bg-brand"
                      />
                      <p className="text-sm font-medium text-gray-800">
                        {TIMELINE_LABELS[entry.status] || entryStatus.label}
                      </p>
                      <p className="text-xs text-gray-500">{formatDateTime(entry.at)}</p>
                      {entry.cancel_reason && (
                        <p className="text-xs text-red-600">{entry.cancel_reason}</p>
                      )}
                    </li>
                  )
                })}
              </ol>
            </div>
          )}
        </section>
      )}

      {status === 2 && (
        <section aria-label="Reseña" className="rounded-2xl border border-gray-100 bg-white p-6 space-y-4">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <h2 className="text-base font-bold text-gray-900">Reseña de la transacción</h2>
            {myReview && <Badge tone="green">Enviada</Badge>}
          </div>

          {myReview ? (
            <div className="rounded-xl bg-gray-50 p-4">
              <StarRating rating={myReview.rating} size={18} showValue />
              {myReview.comment && (
                <p className="mt-2 text-sm text-gray-600">{myReview.comment}</p>
              )}
              <p className="mt-2 text-xs text-gray-400">
                Enviada el {formatDateTime(myReview.created_at)}
              </p>
            </div>
          ) : alreadyReviewedTarget ? (
            <p className="text-sm text-gray-500">
              {counterpartyCompany
                ? 'Ya dejaste una reseña a esta empresa en una transacción anterior.'
                : 'Ya dejaste una reseña a esta parte en una transacción anterior.'}
            </p>
          ) : reviewTarget ? (
            <ReviewForm
              targetLabel={reviewTargetLabel}
              onSubmit={submitReview}
              disabled={busy}
            />
          ) : (
            <p className="text-sm text-gray-500">
              No se pudo identificar a la otra parte de la transacción para dejar la reseña.
            </p>
          )}
        </section>
      )}

      <ConfirmDialog
        open={cancelOpen}
        danger
        title="¿Cancelar la transacción?"
        message="Se libera el monto reservado en tu solicitud y la oferta vuelve a estar activa."
        confirmLabel="Sí, cancelar"
        loading={busy}
        onConfirm={() => {
          if (!reason.trim()) {
            setReasonError('Indica el motivo de la cancelación.')
            return
          }
          run('cancel')
        }}
        onCancel={() => setCancelOpen(false)}
      >
        <label htmlFor="cancel_reason" className="block text-xs font-semibold text-gray-600">
          Motivo
        </label>
        <textarea
          id="cancel_reason"
          rows={3}
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          placeholder="Ej. El agricultor no pudo despachar a tiempo"
          className="mt-1.5 w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
        />
        {reasonError && (
          <p role="alert" className="mt-2 text-sm text-red-600">
            {reasonError}
          </p>
        )}
      </ConfirmDialog>

      <Toast message={toast?.message} tone={toast?.tone} onClose={() => setToast(null)} />
    </div>
  )
}