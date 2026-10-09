import { cn } from '../../lib/cn'
import Button from './Button'
import Icon from './Icon'

export default function ErrorState({
  title = 'Algo salió mal',
  message,
  retryLabel = 'Reintentar',
  onRetry,
  className,
}) {
  return (
    <div role="alert" className={cn('rounded-2xl border border-red-200 bg-red-50 p-6 text-center', className)}>
      <Icon name="error" size={40} className="mx-auto text-red-400" />
      <p className="mt-3 text-sm font-semibold text-red-700">{title}</p>
      {message && <p className="mt-1 text-sm text-red-600">{message}</p>}
      {onRetry && (
        <Button
          variant="danger"
          size="sm"
          onClick={onRetry}
          className="mt-4"
          icon={<Icon name="refresh" size={16} />}
        >
          {retryLabel}
        </Button>
      )}
    </div>
  )
}
