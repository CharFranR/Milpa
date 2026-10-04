import { useCallback, useEffect, useState } from 'react'
import Badge from '../components/ui/Badge'
import Button from '../components/ui/Button'
import EmptyState from '../components/ui/EmptyState'
import ErrorState from '../components/ui/ErrorState'
import Icon from '../components/ui/Icon'
import Skeleton from '../components/ui/Skeleton'
import { formatPrice } from '../lib/format'
import { hasExpired, remainingLabel } from '../lib/liquidations'
import { formatDateTime } from '../lib/supplyStatus'
import { liquidations } from '../services/liquidations'

const REFRESH_MS = 60000

export default function Liquidations() {
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [now, setNow] = useState(() => new Date())

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const data = await liquidations.getOpen()
      setItems(Array.isArray(data) ? data : [])
    } catch (err) {
      setError(err.message || 'No se pudieron cargar las liquidaciones.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  useEffect(() => {
    const timer = setInterval(() => setNow(new Date()), REFRESH_MS)
    return () => clearInterval(timer)
  }, [])

  const visible = items.filter((item) => !hasExpired(item, now))

  return (
    <div className="space-y-6">
      <header>
        <p className="text-xs font-semibold uppercase tracking-wider text-brand">
          Liquidaciones abiertas
        </p>
        <h1 className="mt-1 text-3xl font-bold tracking-tight text-gray-900">
          Lotes de excedente a precio preferencial
        </h1>
        <p className="mt-2 max-w-2xl text-sm text-gray-500">
          Los productores publican aquí sus excedentes de cosecha para venderlos antes de que se
          pierdan. Cada lote indica cuánto falta para que cierre.
        </p>
      </header>

      {loading && (
        <div className="grid gap-5 sm:grid-cols-2 xl:grid-cols-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="rounded-2xl border border-gray-100 bg-white p-5">
              <Skeleton className="h-5 w-40" />
              <Skeleton className="mt-3 h-4 w-24" />
              <Skeleton className="mt-2 h-4 w-32" />
            </div>
          ))}
        </div>
      )}

      {!loading && error && <ErrorState message={error} onRetry={load} />}

      {!loading && !error && visible.length === 0 && (
        <EmptyState
          icon="sell"
          title="No hay liquidaciones abiertas"
          description="Cuando un productor publique un lote de excedente aparecerá en esta lista."
          action={
            <Button onClick={() => (window.location.hash = '#/marketplace')}>
              Ir al marketplace
            </Button>
          }
        />
      )}

      {!loading && !error && visible.length > 0 && (
        <div className="grid gap-5 sm:grid-cols-2 xl:grid-cols-3">
          {visible.map((item) => (
            <article
              key={item.id}
              className="flex flex-col rounded-2xl border border-gray-100 bg-white p-5"
            >
              <div className="flex items-start justify-between gap-3">
                <h2 className="text-base font-semibold text-gray-900">{item.product_name}</h2>
                <Badge tone="green">Abierta</Badge>
              </div>

              <p className="mt-1 text-sm text-gray-500">
                {item.quantity} {item.unit_of_measure}
              </p>

              <div className="mt-4 space-y-1.5 border-t border-gray-100 pt-4 text-sm">
                <p className="flex items-center justify-between">
                  <span className="text-gray-500">Precio total</span>
                  <span className="font-bold text-brand">{formatPrice(item.total_price)}</span>
                </p>
                <p className="flex items-center justify-between">
                  <span className="text-gray-500">Por unidad</span>
                  <span className="font-semibold text-gray-900">
                    {formatPrice(item.unit_price)}
                  </span>
                </p>
                {item.delivery_time && (
                  <p className="flex items-center justify-between">
                    <span className="text-gray-500">Entrega</span>
                    <span className="text-gray-900">{item.delivery_time}</span>
                  </p>
                )}
              </div>

              <div className="mt-4 flex flex-wrap items-center gap-2 border-t border-gray-100 pt-4">
                <Icon name="schedule" size={16} className="text-gray-400" />
                <span className="text-xs text-gray-500">
                  {item.expires_at ? (
                    <>
                      Cierra {formatDateTime(item.expires_at)} ·{' '}
                      <span className="font-semibold text-brand">
                        {remainingLabel(item.expires_at, now)}
                      </span>
                    </>
                  ) : (
                    'Sin fecha de cierre'
                  )}
                </span>
              </div>
            </article>
          ))}
        </div>
      )}
    </div>
  )
}
