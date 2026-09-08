import { cn } from '../lib/cn'
import Icon from './ui/Icon'

export default function SectionHeading({ eyebrow, title, action, centered = false, onDark = false }) {
  return (
    <div className={cn('mb-10', centered && 'text-center')}>
      <span
        className={cn(
          'inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-wider',
          onDark ? 'bg-white/10 text-accent' : 'bg-brand-soft text-brand',
        )}
      >
        <Icon name="eco" size={14} />
        {eyebrow}
      </span>
      <div className={cn('mt-3 flex flex-wrap items-end gap-4', centered && 'justify-center')}>
        <h2 className={cn('text-3xl font-bold tracking-tight sm:text-4xl', onDark ? 'text-white' : 'text-gray-900')}>
          {title}
        </h2>
        {action}
      </div>
    </div>
  )
}