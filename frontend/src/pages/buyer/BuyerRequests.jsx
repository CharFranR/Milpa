import { Link, useNavigate } from 'react-router-dom'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'
import EmptyState from '../../components/ui/EmptyState'
import ErrorState from '../../components/ui/ErrorState'
import Icon from '../../components/ui/Icon'
import Skeleton from '../../components/ui/Skeleton'
import { useSupplyRequests } from '../../hooks/useSupplyRequests'
import { formatPrice } from '../../lib/format'
import { formatDeadline, measurementLabel, requestStatus } from '../../lib/supplyStatus'

function RequestSkeleton() {
  return (
    <div className="space-y-4">
      {[1, 2, 3].map((i) => (
        <div key={i} className="rounded-xl border border-gray-100 bg-white p-5">
          <Skeleton className="h-5 w-48" />
          <Skeleton className="mt-3 h-3 w-72" />
          <Skeleton className="mt-2 h-3 w-40" />
        </div>
      ))}
    </div>
  )
}

export default function BuyerRequests() {
  const navigate = useNavigate()
  const { items, loading, error, reload } = useSupplyRequests()

  return (
    <div className="space-y-6">
      <header className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">
            Mis solicitudes
          </h1>
          <p className="mt-1 text-sm text-gray-500">
            Solicitudes de insumo que publicaste para los agricultores.
          </p>
        </div>
        <Button
          onClick={() => navigate('/dashboard/requests/new')}
          icon={<Icon name="add" size={18} />}
        >
          Nueva solicitud
        </Button>
      </header>

      {loading && <RequestSkeleton />}

      {!loading && error && (
        <ErrorState message={error} onRetry={reload} />
      )}

      {!loading && !error && items.length === 0 && (
        <EmptyState
          icon="inventory_2"
          title="Aún no tienes solicitudes"
          description="Crea una solicitud de insumo para que los agricultores te ofrezcan."
          action={
            <Button onClick={() => navigate('/dashboard/requests/new')}>
              Crear mi primera solicitud
            </Button>
          }
        />
      )}

      {!loading && !error && items.length > 0 && (
        <div className="space-y-4">
          {items.map((item) => {
            const status = requestStatus(item.status)
            return (
              <Link
                key={item.id}
                to={`/dashboard/requests/${item.id}`}
                className="block rounded-xl border border-gray-100 bg-white p-5 transition-colors hover:border-brand/40 hover:shadow-sm"
              >
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div className="min-w-0">
                    <h2 className="truncate font-semibold text-gray-900">
                      {item.product_name}
                    </h2>
                    <p className="mt-1 text-sm text-gray-500">
                      {formatPrice(item.total_amount)} · {measurementLabel(item.unit_of_measure)}
                      {' · '}
                      Disponible: {formatPrice(item.actual_amount)}
                    </p>
                  </div>
                  <Badge tone={status.tone}>{status.label}</Badge>
                </div>

                <div className="mt-3 flex flex-wrap items-center gap-x-5 gap-y-1 text-xs text-gray-500">
                  <span>
                    Pedido: {formatDeadline(item.request_deadline)}
                  </span>
                  <span>
                    Entrega: {formatDeadline(item.delivery_deadline)}
                  </span>
                  <span>
                    {item.multiple_providers
                      ? 'Múltiples agricultores'
                      : 'Un solo agricultor'}
                  </span>
                </div>
              </Link>
            )
          })}
        </div>
      )}
    </div>
  )
}
