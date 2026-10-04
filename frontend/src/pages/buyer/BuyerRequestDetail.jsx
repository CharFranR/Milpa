import { useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'
import ConfirmDialog from '../../components/ui/ConfirmDialog'
import ErrorState from '../../components/ui/ErrorState'
import Icon from '../../components/ui/Icon'
import Skeleton from '../../components/ui/Skeleton'
import Toast from '../../components/ui/Toast'
import { useSupplyRequest } from '../../hooks/useSupplyRequests'
import { formatPrice } from '../../lib/format'
import {
  formatDateTime,
  formatDeadline,
  measurementLabel,
  requestStatus,
} from '../../lib/supplyStatus'
import { supplyRequests } from '../../services/supplyRequests'

function InfoRow({ label, children }) {
  return (
    <div>
      <dt className="text-xs font-semibold uppercase tracking-wide text-gray-400">{label}</dt>
      <dd className="mt-1 text-sm text-gray-800">{children}</dd>
    </div>
  )
}

export default function BuyerRequestDetail() {
  const { id } = useParams()
  const { request, loading, error, reload } = useSupplyRequest(id)
  const [dialog, setDialog] = useState(null)
  const [busy, setBusy] = useState(false)
  const [toast, setToast] = useState(null)

  async function runAction(action) {
    setBusy(true)
    try {
      if (action === 'cancel') {
        await supplyRequests.cancel(id)
      } else {
        await supplyRequests.expire(id)
      }
      setDialog(null)
      setToast({
        message: action === 'cancel' ? 'Solicitud cancelada.' : 'Solicitud marcada como expirada.',
        tone: 'success',
      })
      reload()
    } catch (err) {
      setDialog(null)
      setToast({ message: err.message || 'No se pudo completar la acción.', tone: 'error' })
    } finally {
      setBusy(false)
    }
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-6 w-40" />
        <Skeleton className="h-8 w-72" />
        <div className="rounded-2xl border border-gray-100 bg-white p-6">
          <Skeleton className="h-4 w-full" />
          <Skeleton className="mt-3 h-4 w-2/3" />
          <Skeleton className="mt-3 h-4 w-1/2" />
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
          Volver a mis solicitudes
        </Link>
      </div>
    )
  }

  if (!request) return null

  const status = requestStatus(request.status)
  const isOpen = request.status === 0
  const address = [request.address?.AddressLine, request.address?.Municipality, request.address?.Department]
    .filter(Boolean)
    .join(', ')

  return (
    <div className="space-y-6">
      <header className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <Link
            to="/dashboard"
            className="inline-flex items-center gap-1 text-sm font-semibold text-brand hover:underline"
          >
            <Icon name="chevron_left" size={16} />
            Mis solicitudes
          </Link>
          <div className="mt-2 flex flex-wrap items-center gap-3">
            <h1 className="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">
              {request.product_name}
            </h1>
            <Badge tone={status.tone}>{status.label}</Badge>
          </div>
          <p className="mt-1 text-sm text-gray-500">Creada el {formatDateTime(request.created_at)}</p>
        </div>

        <div className="flex flex-wrap gap-3">
          <Button
            variant="outline"
            onClick={() => setDialog('expire')}
            disabled={!isOpen}
          >
            Marcar expirada
          </Button>
          <Button
            variant="danger"
            onClick={() => setDialog('cancel')}
            disabled={!isOpen}
          >
            Cancelar solicitud
          </Button>
          <Link to={`/dashboard/requests/${id}/edit`}>
            <Button variant="primary" icon={<Icon name="edit" size={16} />}>
              Editar
            </Button>
          </Link>
        </div>
      </header>

      <div className="grid gap-6 md:grid-cols-2">
        <section className="rounded-2xl border border-gray-100 bg-white p-6">
          <h2 className="text-base font-bold text-gray-900">Montos</h2>
          <dl className="mt-4 space-y-4">
            <InfoRow label="Monto total">
              {formatPrice(request.total_amount)} · {measurementLabel(request.unit_of_measure)}
            </InfoRow>
            <InfoRow label="Disponible">
              {formatPrice(request.actual_amount)} de {formatPrice(request.total_amount)}
            </InfoRow>
            <InfoRow label="Proveedores">
              {request.multiple_providers
                ? `Múltiples (mín. ${formatPrice(request.min_amount_per_provider || 0)} por proveedor)`
                : 'Un solo proveedor'}
            </InfoRow>
          </dl>
        </section>

        <section className="rounded-2xl border border-gray-100 bg-white p-6">
          <h2 className="text-base font-bold text-gray-900">Plazos y entrega</h2>
          <dl className="mt-4 space-y-4">
            <InfoRow label="Fecha límite de solicitud">
              {formatDeadline(request.request_deadline)}
            </InfoRow>
            <InfoRow label="Fecha límite de entrega">
              {formatDeadline(request.delivery_deadline)}
            </InfoRow>
            <InfoRow label="Dirección">{address || 'Sin dirección registrada'}</InfoRow>
          </dl>
        </section>

        <section className="rounded-2xl border border-gray-100 bg-white p-6 md:col-span-2">
          <h2 className="text-base font-bold text-gray-900">Detalles</h2>
          <p className="mt-3 whitespace-pre-wrap text-sm text-gray-600">
            {request.description || 'Sin descripción adicional.'}
          </p>
        </section>
      </div>

      <ConfirmDialog
        open={dialog === 'cancel'}
        danger
        title="¿Cancelar la solicitud?"
        message="Dejará de estar visible para los proveedores. Esta acción no se puede deshacer."
        confirmLabel="Sí, cancelar"
        loading={busy}
        onConfirm={() => runAction('cancel')}
        onCancel={() => setDialog(null)}
      />

      <ConfirmDialog
        open={dialog === 'expire'}
        danger
        title="¿Marcar como expirada?"
        message="La solicitud quedará cerrada por tiempo, igual que si se hubiera vencido."
        confirmLabel="Sí, expirar"
        loading={busy}
        onConfirm={() => runAction('expire')}
        onCancel={() => setDialog(null)}
      />

      <Toast message={toast?.message} tone={toast?.tone} onClose={() => setToast(null)} />
    </div>
  )
}
