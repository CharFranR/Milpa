import { cn } from '../lib/cn'

const SIZES = {
  xs: 'h-7 w-7 text-xs',
  sm: 'h-9 w-9 text-sm',
  md: 'h-11 w-11 text-base',
  lg: 'h-14 w-14 text-lg',
}

const TONE_CLASSES = {
  'brand-soft': 'bg-brand-soft text-brand',
  brand: 'bg-brand text-white',
  accent: 'bg-accent text-night',
  gray: 'bg-gray-200 text-gray-700',
}

export default function Avatar({
  initials,
  name,
  size = 'md',
  tone = 'brand-soft',
  className,
}) {
  return (
    <span
      role="img"
      aria-label={name ? `Foto o iniciales de ${name}` : undefined}
      className={cn(
        'inline-flex shrink-0 items-center justify-center rounded-full font-semibold',
        SIZES[size],
        TONE_CLASSES[tone],
        className,
      )}
    >
      {initials}
    </span>
  )
}