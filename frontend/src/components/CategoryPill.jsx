import { cn } from '../lib/cn'
import Icon from './ui/Icon'

export default function CategoryPill({ icon, name, count, active, className, onClick }) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={active}
      className={cn(
        'flex items-center gap-2.5 rounded-2xl border px-4 py-3 text-sm font-medium transition-colors',
        'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand',
        active
          ? 'border-brand bg-brand text-white'
          : 'border-gray-200 bg-white text-gray-700 hover:border-brand/40 hover:bg-brand-soft',
        className,
      )}
    >
      <Icon name={icon} size={20} className={active ? 'text-white' : 'text-brand'} />
      <span className="font-semibold">{name}</span>
      <span className={cn('ml-auto text-xs', active ? 'text-white/80' : 'text-gray-400')}>
        {count}
      </span>
    </button>
  )
}