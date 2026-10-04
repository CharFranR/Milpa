import { useEffect, useState } from 'react'
import Badge from '../ui/Badge'
import EmptyState from '../ui/EmptyState'
import ErrorState from '../ui/ErrorState'
import Skeleton from '../ui/Skeleton'
import { useOffersByRequest } from '../../hooks/useSupplyOffers'
import { formatPrice } from '../../lib/format'
import { formatDateTime, measurementLabel, offerStatus } from '../../lib/supplyStatus'
import { users } from '../../services/api'

function shortId(id) {
  return id ? String(id).slice(0, 8) : '—'
}

export default function OffersSection({ requestId }) {
  const { items, loading, error, reload } = useOffersByRequest(requestId)
  const [names, setNames] = useState({})

  useEffect(() => {
    const ids = [...new Set(items.map((item) => item.supplier_id).filter(Boolean))]
    if (ids.length === 0) return undefined
    let alive = true
    Promise.all(
      ids.map((userId) =>
        users
          .getById(userId)
          .then((user) => [userId, `${user.first_name} ${user.last_name}`.trim()])
          .catch(() => [userId, '']),
      ),
    ).then((pairs) => {
      if (alive) setNames(Object.fromEntries(pairs))
    })
    return () => {
      alive = false
    }
  }, [items])

  return (
    <section aria-label="Ofertas recibidas" className="space-y-4">
      <div className="flex items-baseline justify-between gap-3">
        <h2 className="text-base font-bold text-gray-900">Ofertas recibidas</h2>
        {!loading && !error && items.length > 0 && (
          <span className="text-xs text-gray-400">{items.length} en total</span>
        )}
      </div>

      {loading && (
        <div className="space-y-4">
          {[1, 2].map((i) => (
            <div key={i} className="rounded-xl border border-gray-100 bg-white p-5">
              <Skeleton className="h-5 w-56" />
              <Skeleton className="mt-3 h-3 w-72" />
            </div>
          ))}
        </div>
      )}

      {!loading && error && <ErrorState message={error} onRetry={reload} />}

      {!loading && !error && items.length === 0 && (
        <EmptyState
          icon="handshake"
          title="Sin ofertas todavía"
          description="Los proveedores verán tu solicitud en «Solicitudes disponibles» y podrán ofertar."
        />
      )}

      {!loading && !error && items.length > 0 && (
        <div className="space-y-4">
          {items.map((offer) => {
            const status = offerStatus(offer.status)
            return (
              <article
                key={offer.id}
                className="rounded-xl border border-gray-100 bg-white p-5 transition-shadow hover:shadow-sm"
              >
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div className="min-w-0">
                    <h3 className="font-semibold text-gray-900">
                      {names[offer.supplier_id] || `Proveedor ${shortId(offer.supplier_id)}`}
                    </h3>
                    <p className="mt-1 text-sm text-gray-500">
                      {formatPrice(offer.total_amount)} · {formatPrice(offer.price_per_unit)} por{' '}
                      {measurementLabel(offer.measurement)}
                      {offer.delivery_day
                        ? ` · Entrega ${formatDateTime(offer.delivery_day)}`
                        : ' · Sin fecha de entrega'}
                    </p>
                  </div>
                  <Badge tone={status.tone}>{status.label}</Badge>
                </div>

                {offer.comments && (
                  <p className="mt-3 text-sm text-gray-600">{offer.comments}</p>
                )}
              </article>
            )
          })}
        </div>
      )}
    </section>
  )
}
