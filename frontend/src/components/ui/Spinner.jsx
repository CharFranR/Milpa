import { cn } from '../../lib/cn'
import Icon from './Icon'

export default function Spinner({ size = 20, label, className }) {
  return (
    <span role="status" className={cn('inline-flex items-center gap-2', className)}>
      <Icon name="progress_activity" size={size} className="animate-spin" />
      {label && <span className="text-sm font-medium text-gray-600">{label}</span>}
    </span>
  )
}
