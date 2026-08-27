import { cn } from '../../lib/cn'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'
import Avatar from '../../components/Avatar'
import { producerRequests } from '../../mocks/producer'

export default function ProducerRequests() {
  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold text-gray-900">Solicitudes de compra</h1>

      <div className="space-y-4">
        {producerRequests.map((req) => (
          <article
            key={req.id}
            className={cn(
              'rounded-xl border border-gray-100 bg-white p-5 transition-colors',
              req.status === 'pending' ? 'border-amber-200' : 'border-gray-100',
            )}
          >
            <div className="flex items-start justify-between gap-4">
              <div className="flex items-center gap-3 min-w-0 flex-1">
                <Avatar
                  initials={req.buyer.avatar}
                  name={req.buyer.name}
                  size="md"
                  tone={req.buyer.tone}
                />
                <div className="min-w-0">
                  <p className="font-semibold text-gray-900 truncate">{req.buyer.name}</p>
                  <p className="text-sm text-gray-500">{req.product} · {req.qty}</p>
                </div>
              </div>

              <div className="flex items-center gap-3 shrink-0">
                <Badge
                  tone={req.status === 'pending' ? 'amber' : 'brand'}
                  className="shrink-0"
                >
                  {req.status === 'pending' ? 'Pendiente' : 'Respondido'}
                </Badge>
                <span className="text-sm text-gray-500 whitespace-nowrap shrink-0">{req.time}</span>
              </div>
            </div>

            <p className="mt-3 text-sm text-gray-600 italic">"{req.message}"</p>

            {req.status === 'pending' && (
              <div className="mt-4 flex items-center justify-end gap-2">
                <Button variant="outline" size="sm">
                  Ver producto
                </Button>
                <Button variant="primary" size="sm">
                  Responder
                </Button>
              </div>
            )}
          </article>
        ))}
      </div>
    </div>
  )
}