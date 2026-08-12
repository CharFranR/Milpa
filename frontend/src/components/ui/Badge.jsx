import { cn } from '../../lib/cn'

const TONES = {
  brand: 'bg-brand-soft text-brand',
  accent: 'bg-accent-soft text-brand',
  red: 'bg-red-100 text-red-700',
  amber: 'bg-amber-100 text-amber-700',
  gray: 'bg-gray-200 text-gray-600',
  green: 'bg-green-100 text-green-700',
}

export default function Badge({ children, tone = 'brand', className }) {
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-semibold',
        TONES[tone],
        className,
      )}
    >
      {children}
    </span>
  )
}