import Icon from './ui/Icon'
import { cn } from '../lib/cn'

const TONE_CLASSES = ['bg-red-100 text-red-600', 'bg-blue-100 text-blue-600', 'bg-amber-100 text-amber-700', 'bg-brand-soft text-brand']

export default function StatCard({ icon, value, label, index = 0, className }) {
  const tone = TONE_CLASSES[index % TONE_CLASSES.length]
  return (
    <div className={cn('rounded-2xl border border-gray-100 bg-white p-5 shadow-sm', className)}>
      <div className="flex items-start justify-between gap-3">
        <div>
          <p className="text-2xl font-bold tracking-tight text-gray-900">{value}</p>
          <p className="mt-0.5 text-sm text-gray-500">{label}</p>
        </div>
        <span className={cn('inline-flex h-10 w-10 items-center justify-center rounded-xl', tone)}>
          <Icon name={icon} size={20} />
        </span>
      </div>
    </div>
  )
}