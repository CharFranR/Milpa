import { useCallback, useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import Badge from '../../components/ui/Badge'
import ErrorState from '../../components/ui/ErrorState'
import Icon from '../../components/ui/Icon'
import Skeleton from '../../components/ui/Skeleton'
import { formatPrice } from '../../lib/format'
import {
  formatDateTime,
  matchStatus,
  measurementLabel,
  transactionStatus,
} from '../../lib/supplyStatus'
import { matches } from '../../services/matches'
import { transactions } from '../../services/transactions'

export default function BuyerMatchDetail() {
  const { matchId } = useParams()
  const [match, setMatch] = useState(null)
  const [transaction, setTransaction] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const reload = useCallback(() => {
    if (!matchId) return
    setLoading(true)
    setError('')
    Promise.all([matches.getById(matchId), transactions.getByMatch(matchId)])
      .then(([matchData, transactionData]) => {
        setMatch(matchData)
        setTransaction(transactionData)
      })
      .catch((err) => setError(err.message || 'No se pudo cargar el match.'))
      .finally(() => setLoading(false))
  }, [matchId])

  useEffect(() => {
    reload()
  }, [reload])

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
          to="/dashboard"
          className="inline-flex items-center gap-1 text-sm font-semibold text-brand hover:underline"
        >
          <Icon name="chevron_left" size={16} />
          Mis solicitudes
        </Link>
      </div>
    )
  }

  if (!match) return null

  const mStatus = matchStatus(match.status)
  const tStatus = transaction ? transactionStatus(transaction.status) : null

  return (
    <div className="space-y-6">
      <header>
        <Link
          to={`/dashboard/requests/${match.supply_request}`}
          className="inline-flex items-center gap-1 text-sm font-semibold text-brand hover:underline"
        >
          <Icon name="chevron_left" size={16} />
          Volver a la solicitud
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
          Reservado de tu solicitud. El proveedor y tú ya pueden coordinar por chat.
        </p>
      </section>

      {transaction && (
        <section className="rounded-2xl border border-gray-100 bg-white p-6">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <h2 className="text-base font-bold text-gray-900">Transacción</h2>
            <Badge tone={tStatus.tone}>{tStatus.label}</Badge>
          </div>
          <dl className="mt-4 space-y-3">
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wide text-gray-400">
                Identificador
              </dt>
              <dd className="mt-1 text-sm text-gray-800">{String(transaction.id).slice(0, 8)}</dd>
            </div>
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wide text-gray-400">
                Creada
              </dt>
              <dd className="mt-1 text-sm text-gray-800">
                {formatDateTime(transaction.created_at)}
              </dd>
            </div>
          </dl>
        </section>
      )}
    </div>
  )
}
