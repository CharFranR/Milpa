import { cn } from '../../lib/cn'

export default function Icon({ name, size = 20, filled = false, weight = 400, className }) {
  return (
    <span
      aria-hidden="true"
      className={cn('material-symbols-rounded inline-flex shrink-0 select-none', className)}
      style={{
        fontSize: size,
        fontVariationSettings: `'FILL' ${filled ? 1 : 0}, 'wght' ${weight}`,
      }}
    >
      {name}
    </span>
  )
}