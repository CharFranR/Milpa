import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import Badge from '../ui/Badge'
import Button from '../ui/Button'
import ConfirmDialog from '../ui/ConfirmDialog'
import EmptyState from '../ui/EmptyState'
import ErrorState from '../ui/ErrorState'
import Icon from '../ui/Icon'
import Skeleton from '../ui/Skeleton'
import Toast from '../ui/Toast'
import { usePrioritizedOffers } from '../../hooks/useMatches'
import { formatPrice, money } from '../../lib/format'
import { formatDateTime, measurementLabel } from '../../lib/supplyStatus'
import { matches } from '../../services/matches'
import { users } from '../../services/api'

function shortId(id) {
  return id ? String(id).slice(0, 8) : '—'
}

function formatScore(value) {
  return typeof value === 'number' ? value.toFixed(2) : '—'
}

export default function OffersSection({ requestId }) {
  const navigate = useNavigate()
  const { items, loading, error, reload } = usePrioritizedOffers(requestId)
  const [names, setNames] = useState({})
  const [busyId, setBusyId] = useState(null)
  const [dialog, setDialog] = useState(null)
  const [toast, setToast] = useState(null)

  useEffect(() => {
    const ids = [...new Set(items.map((item) => item.offer?.supplier_id).filter(Boolean))]
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

  async function handleLike(offer) {
    setBusyId(offer.id)
    try {
      const created = await matches.like(offer.id)
      navigate(`/dashboard/matches/${created.match.id}`)
    } catch (err) {
      setToast({ message: err.message || 'No se pudo crear el match.', tone: 'error' })
      setBusyId(null)
    }
  }

  async function handlePass() {
    const offer = dialog
    setBusyId(offer.id)
    try {
      await matches.pass(offer.id)
      setDialog(null)
      setToast({ message: 'Oferta descartada.', tone: 'success' })
      reload()
    } catch (err) {
      setDialog(null)
      setToast({ message: err.message || 'No se pudo descartar la oferta.', tone: 'error' })
    } finally {
      setBusyId(null)
    }
  }

  return (
    <section aria-label="Ofertas recibidas" className="space-y-4">
      <div className="flex items-baseline justify-between gap-3">
        <div>
          <h2 className="text-base font-bold text-gray-900">Ofertas recibidas</h2>
          <p className="text-xs text-gray-400">Ordenadas por recomendación</p>
        </div>
        {!loading && !error && items.length > 0 && (
          <span className="text-xs text-gray-400">{items.length} activas</span>
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
              description="Los agricultores verán tu solicitud en «Solicitudes disponibles» y podrán ofertar."
            />
      )}

      {!loading && !error && items.length > 0 && (
        <div className="space-y-4">
          {items.map(({ offer, score, available_quantity: available, contributions = [] }) => (
            <article
              key={offer.id}
              className="rounded-xl border border-gray-100 bg-white p-5 transition-shadow hover:shadow-sm"
            >
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div className="min-w-0">
<h3 className="font-semibold text-gray-900">
                      {names[offer.supplier_id] || `Agricultor ${shortId(offer.supplier_id)}`}
                    </h3>
                  <p className="mt-1 text-sm text-gray-500">
                    {formatPrice(offer.total_amount)} · {formatPrice(offer.price_per_unit)} por{' '}
                    {measurementLabel(offer.measurement)}
                    {offer.delivery_day
                      ? ` · Entrega ${formatDateTime(offer.delivery_day)}`
                      : ' · Sin fecha de entrega'}
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <Badge tone="brand">Puntaje {formatScore(score)}</Badge>
                </div>
              </div>

              <p className="mt-3 text-sm text-gray-600">
                Disponibilidad del agricultor: {money(available)} {measurementLabel(offer.measurement)}
              </p>

              {offer.comments && (
                <p className="mt-2 whitespace-pre-wrap text-sm text-gray-600">{offer.comments}</p>
              )}

              {contributions.length > 0 && (
                <ul className="mt-3 flex flex-wrap gap-2">
                  {contributions.map((c) => (
                    <li
                      key={c.factor}
                      className="rounded-full bg-gray-50 px-2.5 py-1 text-xs text-gray-500"
                      title={`Peso ${(c.weight * 100).toFixed(0)}% · ${c.factor}`}
                    >
                      {c.factor} · {formatScore(c.weighted_score)}
                    </li>
                  ))}
                </ul>
              )}

              <div className="mt-4 flex flex-wrap gap-3">
                <Button
                  onClick={() => handleLike(offer)}
                  disabled={busyId === offer.id}
                  icon={<Icon name="handshake" size={16} />}
                >
                  {busyId === offer.id ? 'Procesando…' : 'Me interesa'}
                </Button>
                <Button
                  variant="outline"
                  onClick={() => setDialog(offer)}
                  disabled={busyId === offer.id}
                >
                  Descartar
                </Button>
              </div>
            </article>
          ))}
        </div>
      )}

      <ConfirmDialog
        open={Boolean(dialog)}
        danger
        title="¿Descartar esta oferta?"
        message="El agricultor la verá como rechazada y no podrás seleccionarla después."
        confirmLabel="Sí, descartar"
        loading={busyId === dialog?.id}
        onConfirm={handlePass}
        onCancel={() => setDialog(null)}
      />

      <Toast message={toast?.message} tone={toast?.tone} onClose={() => setToast(null)} />
    </section>
  )
}
