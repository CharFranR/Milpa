import { cn } from '../../lib/cn'
import Button from '../ui/Button'

export default function ProducerCard({
  producer,
  className,
}) {
  if (!producer) return null

  const initials = producer.name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()

  const locationParts = producer.location ? producer.location.split(', ') : []
  const city = locationParts[0] || producer.city || ''
  const region = locationParts[1] || producer.region || ''

  return (
    <article className={cn('rounded-2xl border border-gray-100 bg-white p-5 shadow-sm hover:shadow-md transition-shadow min-w-0', className)}>
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
            {producer.active && (
              <span className="inline-flex h-2 w-2 rounded-full bg-green-500" aria-label="Activo" />
            )}
          </div>
          <p className="mt-0.5 text-sm text-gray-500 truncate">{producer.farm}</p>
          <p className="text-sm text-gray-500">
            <span className="inline-flex items-center gap-1">
              <span className="text-[10px] leading-none">📍</span>
              {city}{region && `, ${region}`}
            </span>
          </p>
        </div>
      </div>

      <div className="mt-4 flex flex-wrap items-center justify-between gap-3 border-t border-gray-100 pt-4">
        <div className="flex flex-wrap items-center gap-3 text-sm text-gray-500">
          <span className="flex items-center gap-1 whitespace-nowrap">
            <span className="text-[10px] leading-none">📦</span>
            {producer.productsCount} productos
          </span>
          <span className="flex items-center gap-1 whitespace-nowrap">
            <span className="text-[10px] leading-none">⭐</span>
            {producer.rating}
          </span>
          <span className="flex items-center gap-1 whitespace-nowrap">
            <span className="text-[10px] leading-none">📅</span>
            Desde {producer.since}
          </span>
        </div>

        <div className="flex items-center gap-2 flex-wrap shrink-0">
          <Button variant="outline" size="sm">
            Ver perfil
          </Button>
          <Button variant="primary" size="sm">
            Gestionar
          </Button>
        </div>
      </div>
    </article>
  )
}