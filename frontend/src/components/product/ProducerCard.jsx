import { cn } from '../../lib/cn'
import Icon from '../ui/Icon'
import StarRating from '../StarRating'
import Button from '../ui/Button'
import { producerById } from '../../mocks/catalog'

export default function ProducerCard({
  producerId,
  showActions = true,
  className,
}) {
  const producer = producerById(producerId)

  if (!producer) return null

  const initials = producer.name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()

  return (
    <article className={cn('rounded-2xl border border-gray-100 bg-white p-5 shadow-sm hover:shadow-md transition-shadow', className)}>
      <div className="flex items-center gap-4">
        <span
          className="flex h-16 w-16 shrink-0 items-center justify-center rounded-full bg-brand-soft text-brand text-xl font-bold"
          role="img"
          aria-label={producer.name}
        >
          {initials}
        </span>
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <h3 className="font-semibold text-gray-900 truncate">{producer.name}</h3>
            <span className="inline-flex h-2 w-2 rounded-full bg-green-500" aria-label="Activo" />
          </div>
          <p className="mt-0.5 text-sm text-gray-500 truncate">{producer.farm}</p>
          <p className="text-sm text-gray-500">
            <Icon name="location_on" size={14} className="inline-block mr-1" />
            {producer.city}, {producer.region}
          </p>
        </div>
      </div>

      <div className="mt-4 flex items-center justify-between border-t border-gray-100 pt-4">
        <div className="flex items-center gap-3 text-sm text-gray-500">
          <span className="flex items-center gap-1">
            <Icon name="inventory_2" size={14} />
            {producer.productsCount} productos
          </span>
          <span className="flex items-center gap-1">
            <StarRating rating={producer.rating} size={12} showValue />
          </span>
          <span className="flex items-center gap-1">
            <Icon name="calendar_today" size={14} />
            Desde {producer.since}
          </span>
        </div>

        {showActions && (
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm">
              Ver perfil
            </Button>
            <Button variant="primary" size="sm">
              Gestionar
            </Button>
          </div>
        )}
      </div>
    </article>
  )
}