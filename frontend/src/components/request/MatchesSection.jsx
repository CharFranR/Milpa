import { Link } from 'react-router-dom'
import Badge from '../ui/Badge'
import EmptyState from '../ui/EmptyState'
import ErrorState from '../ui/ErrorState'
import Icon from '../ui/Icon'
import Skeleton from '../ui/Skeleton'
import { useRequestMatches } from '../../hooks/useMatches'
import { formatPrice } from '../../lib/format'
import { formatDateTime, matchStatus, measurementLabel } from '../../lib/supplyStatus'

export default function MatchesSection({ requestId }) {
  const { items, loading, error, reload } = useRequestMatches(requestId)

  return (
    <section aria-label="Matches" className="space-y-4">
      <div className="flex items-baseline justify-between gap-3">
        <div>
          <h2 className="text-base font-bold text-gray-900">Matches</h2>
          <p className="text-xs text-gray-400">Ofertas que ya reservaron monto de tu solicitud</p>
        </div>
        {!loading && !error && items.length > 0 && (
          <span className="text-xs text-gray-400">{items.length}</span>
        )}
      </div>

      {loading && (
        <div className="space-y-3">
          {[1, 2].map((i) => (
            <div key={i} className="rounded-xl border border-gray-100 bg-white p-5">
              <Skeleton className="h-5 w-64" />
            </div>
          ))}
        </div>
      )}

      {!loading && error && <ErrorState message={error} onRetry={reload} />}

      {!loading && !error && items.length === 0 && (
        <EmptyState
          icon="bolt"
          title="Sin matches"
          description="Cuando te interese una oferta y confirmes, el match aparecerá aquí con su transacción."
        />
      )}

      {!loading && !error && items.length > 0 && (
        <ul className="space-y-3">
          {items.map((match) => {
            const status = matchStatus(match.status)
            return (
              <li
                key={match.id}
                className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-gray-100 bg-white p-4"
              >
                <div className="min-w-0">
                  <p className="text-sm font-semibold text-gray-900">
                    {formatPrice(match.matched_amount)} {measurementLabel(match.amount_unit)}
                  </p>
                  <p className="mt-0.5 text-xs text-gray-500">
                    Creado el {formatDateTime(match.created_at)} · {String(match.id).slice(0, 8)}
                  </p>
                </div>
                <div className="flex items-center gap-3">
                  <Badge tone={status.tone}>{status.label}</Badge>
                  <Link
                    to={`/dashboard/matches/${match.id}`}
                    className="inline-flex items-center gap-1 text-sm font-semibold text-brand hover:underline"
                  >
                    Ver detalle
                    <Icon name="chevron_right" size={16} />
                  </Link>
                </div>
              </li>
            )
          })}
        </ul>
      )}
    </section>
  )
}
