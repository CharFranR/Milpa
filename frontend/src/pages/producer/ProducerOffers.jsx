import { useEffect, useState } from 'react'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'
import ConfirmDialog from '../../components/ui/ConfirmDialog'
import EmptyState from '../../components/ui/EmptyState'
import ErrorState from '../../components/ui/ErrorState'
import Icon from '../../components/ui/Icon'
import Skeleton from '../../components/ui/Skeleton'
import Toast from '../../components/ui/Toast'
import { useMyOffers } from '../../hooks/useSupplyOffers'
import { formatPrice } from '../../lib/format'
import { formatDateTime, measurementLabel, offerStatus } from '../../lib/supplyStatus'
import { supplyOffers } from '../../services/supplyOffers'
import { supplyRequests } from '../../services/supplyRequests'

function shortId(id) {
  return id ? String(id).slice(0, 8) : '—'
}

export default function ProducerOffers() {
  const { items, loading, error, reload } = useMyOffers()
  const [names, setNames] = useState({})
  const [dialog, setDialog] = useState(null)
  const [busy, setBusy] = useState(false)
  const [toast, setToast] = useState(null)

  useEffect(() => {
    const ids = [...new Set(items.map((item) => item.supply_request_id).filter(Boolean))]
    if (ids.length === 0) return undefined
    let alive = true
    Promise.all(
      ids.map((requestId) =>
        supplyRequests
          .getById(requestId)
          .then((request) => [requestId, request.product_name])
          .catch(() => [requestId, '']),
      ),
    ).then((pairs) => {
      if (alive) setNames(Object.fromEntries(pairs))
    })
    return () => {
      alive = false
    }
  }, [items])

  async function handleWithdraw(offer) {
    setBusy(true)
    try {
      await supplyOffers.withdraw(offer.id)
      setDialog(null)
      setToast({ message: 'Oferta retirada.', tone: 'success' })
      reload()
    } catch (err) {
      setDialog(null)
      setToast({ message: err.message || 'No se pudo retirar la oferta.', tone: 'error' })
    } finally {
      setBusy(false)
    }
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-48" />
        <div className="space-y-4">
          {[1, 2].map((i) => (
            <div key={i} className="rounded-xl border border-gray-100 bg-white p-5">
              <Skeleton className="h-5 w-56" />
              <Skeleton className="mt-3 h-3 w-72" />
            </div>
          ))}
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-6">
        <ErrorState message={error} onRetry={reload} />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <header>
        <h1 className="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">Mis ofertas</h1>
        <p className="mt-1 text-sm text-gray-500">
          Ofertas que enviaste a las solicitudes de los compradores.
        </p>
      </header>

      {items.length === 0 ? (
        <EmptyState
          icon="handshake"
          title="Todavía no has ofrecido"
          description="Revisa las solicitudes disponibles y envía tu primera oferta."
        />
      ) : (
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
                    <h2 className="font-semibold text-gray-900">
                      {names[offer.supply_request_id] || `Solicitud ${shortId(offer.supply_request_id)}`}
                    </h2>
                    <p className="mt-1 text-sm text-gray-500">
                      {formatPrice(offer.total_amount)} · {formatPrice(offer.price_per_unit)} por{' '}
                      {measurementLabel(offer.measurement)}
                      {offer.delivery_day ? ` · Entrega ${formatDateTime(offer.delivery_day)}` : ''}
                    </p>
                    <time className="mt-1 block text-xs text-gray-400">
                      Enviada el {formatDateTime(offer.created_at)}
                    </time>
                  </div>

                  <div className="flex shrink-0 items-center gap-2">
                    <Badge tone={status.tone}>{status.label}</Badge>
                    {offer.status === 0 && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setDialog(offer)}
                        disabled={busy}
                      >
                        Retirar
                      </Button>
                    )}
                  </div>
                </div>

                {offer.comments && (
                  <p className="mt-3 text-sm text-gray-600">{offer.comments}</p>
                )}
              </article>
            )
          })}
        </div>
      )}

      <ConfirmDialog
        open={Boolean(dialog)}
        danger
        title="¿Retirar la oferta?"
        message="El comprador dejará de verla. Es definitivo: no podrás volver a ofertar en esta misma solicitud."
        confirmLabel="Sí, retirar"
        loading={busy}
        onConfirm={() => handleWithdraw(dialog)}
        onCancel={() => setDialog(null)}
      />

      <Toast message={toast?.message} tone={toast?.tone} onClose={() => setToast(null)} />

      <p className="flex items-center gap-2 text-xs text-gray-400">
        <Icon name="info" size={14} />
        Las ofertas retiradas se conservan en el historial del comprador como rechazadas.
      </p>
    </div>
  )
}
