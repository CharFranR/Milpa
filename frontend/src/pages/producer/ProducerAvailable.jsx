import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'
import EmptyState from '../../components/ui/EmptyState'
import ErrorState from '../../components/ui/ErrorState'
import Icon from '../../components/ui/Icon'
import Skeleton from '../../components/ui/Skeleton'
import { useAvailableRequests } from '../../hooks/useSupplyOffers'
import { formatPrice } from '../../lib/format'
import { formatDeadline, measurementLabel, requestStatus } from '../../lib/supplyStatus'

export default function ProducerAvailable() {
  const navigate = useNavigate()
  const { items, loading, error, reload } = useAvailableRequests()
  const [query, setQuery] = useState('')

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q) return items
    return items.filter((item) =>
      [item.product_name, item.description, item.address?.Department, item.address?.Municipality]
        .filter(Boolean)
        .some((value) => String(value).toLowerCase().includes(q)),
    )
  }, [items, query])

  return (
    <div className="space-y-6">
      <header className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">
            Solicitudes disponibles
          </h1>
          <p className="mt-1 text-sm text-gray-500">
            Insumos que los compradores están pidiendo y donde aún puedes ofertar.
          </p>
        </div>

        <div className="relative w-full sm:w-72">
          <Icon
            name="search"
            size={18}
            className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"
          />
          <input
            type="search"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Buscar por producto o departamento"
            aria-label="Buscar solicitudes"
            className="w-full rounded-full border border-gray-300 bg-white py-2.5 pl-10 pr-4 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
          />
        </div>
      </header>

      {loading && (
        <div className="space-y-4">
          {[1, 2, 3].map((i) => (
            <div key={i} className="rounded-xl border border-gray-100 bg-white p-5">
              <Skeleton className="h-5 w-52" />
              <Skeleton className="mt-3 h-3 w-72" />
              <Skeleton className="mt-2 h-3 w-44" />
            </div>
          ))}
        </div>
      )}

      {!loading && error && <ErrorState message={error} onRetry={reload} />}

      {!loading && !error && items.length === 0 && (
        <EmptyState
          icon="inbox"
          title="No hay solicitudes abiertas"
          description="Cuando un comprador publique una solicitud de insumo aparecerá aquí."
        />
      )}

      {!loading && !error && items.length > 0 && filtered.length === 0 && (
        <EmptyState
          icon="search_off"
          title="Sin resultados"
          description={`Ninguna solicitud coincide con «${query.trim()}».`}
        />
      )}

      {!loading && !error && filtered.length > 0 && (
        <div className="space-y-4">
          {filtered.map((item) => {
            const status = requestStatus(item.status)
            return (
              <article
                key={item.id}
                className="rounded-xl border border-gray-100 bg-white p-5 transition-shadow hover:shadow-sm"
              >
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div className="min-w-0">
                    <h2 className="font-semibold text-gray-900">{item.product_name}</h2>
                    <p className="mt-1 text-sm text-gray-500">
                      {formatPrice(item.total_amount)} · {measurementLabel(item.unit_of_measure)}
                      {' · '}
                      Disponible: {formatPrice(item.actual_amount)}
                    </p>
                  </div>
                  <div className="flex shrink-0 items-center gap-2">
                    <Badge tone={status.tone}>{status.label}</Badge>
                    <Button
                      size="sm"
                      onClick={() => navigate(`/producer/available/${item.id}/offer`)}
                    >
                      Ofrecer
                    </Button>
                  </div>
                </div>

                <div className="mt-3 flex flex-wrap items-center gap-x-5 gap-y-1 text-xs text-gray-500">
                  <span>Entrega hasta: {formatDeadline(item.delivery_deadline)}</span>
                  <span>{item.address?.Department || 'Sin departamento'}</span>
<span>
                      {item.multiple_providers
                        ? `Mín. ${formatPrice(item.min_amount_per_provider || 0)} por agricultor`
                        : 'Un solo agricultor'}
                    </span>
                </div>

                {item.description && (
                  <p className="mt-3 line-clamp-2 text-sm text-gray-600">{item.description}</p>
                )}
              </article>
            )
          })}
        </div>
      )}
    </div>
  )
}
