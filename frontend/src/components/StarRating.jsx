import { cn } from '../lib/cn'
import Icon from './ui/Icon'

export default function StarRating({ rating, reviews, size = 16, className, showValue = false }) {
  const full = Math.floor(rating)
  const hasHalf = rating - full >= 0.5

  return (
    <span className={cn('inline-flex items-center gap-1', className)} role="img" aria-label={`Valoración ${rating} de 5`}>
      <span className="inline-flex items-center text-amber-400">
        {Array.from({ length: 5 }, (_, i) => {
          const name = i < full ? 'star' : i === full && hasHalf ? 'star_half' : 'star'
          return <Icon key={i} name={name} size={size} filled={i < full} weight={500} />
        })}
      </span>
      {showValue && (
        <span className="ml-0.5 text-xs font-semibold text-gray-700">{rating}</span>
      )}
      {reviews !== undefined && (
        <span className="text-xs text-gray-500">({reviews} opiniones)</span>
      )}
    </span>
  )
}