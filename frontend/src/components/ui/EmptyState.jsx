import { cn } from '../../lib/cn'
import Icon from './Icon'

export default function EmptyState({ icon = 'inbox', title, description, action, className }) {
  return (
    <div className={cn('rounded-2xl border border-gray-100 bg-white p-10 text-center', className)}>
      <Icon name={icon} size={40} className="mx-auto text-gray-300" />
      <p className="mt-3 text-sm font-semibold text-gray-700">{title}</p>
      {description && <p className="mt-1 text-sm text-gray-500">{description}</p>}
      {action && <div className="mt-4 flex justify-center">{action}</div>}
    </div>
  )
}
