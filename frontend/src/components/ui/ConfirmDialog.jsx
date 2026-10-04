import { useEffect, useRef } from 'react'
import { cn } from '../../lib/cn'
import Button from './Button'
import Icon from './Icon'

export default function ConfirmDialog({
  open,
  title = '¿Estás seguro?',
  message,
  confirmLabel = 'Confirmar',
  cancelLabel = 'Cancelar',
  danger = false,
  loading = false,
  onConfirm,
  onCancel,
  className,
}) {
  const cancelRef = useRef(null)

  useEffect(() => {
    if (!open) return undefined
    function handleKeyDown(event) {
      if (event.key === 'Escape') onCancel?.()
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [open, onCancel])

  useEffect(() => {
    if (open) cancelRef.current?.focus()
  }, [open])

  if (!open) return null

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 backdrop-blur-sm"
      role="dialog"
      aria-modal="true"
      aria-labelledby="confirm-dialog-title"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) onCancel?.()
      }}
    >
      <div className={cn('w-full max-w-sm rounded-2xl bg-white p-6 shadow-xl', className)}>
        <div className="flex items-start gap-3">
          <span className={cn('mt-0.5 shrink-0', danger ? 'text-red-600' : 'text-brand')}>
            <Icon name={danger ? 'error' : 'help'} size={24} />
          </span>
          <div>
            <h2 id="confirm-dialog-title" className="text-base font-semibold text-gray-900">
              {title}
            </h2>
            {message && <p className="mt-1 text-sm text-gray-600">{message}</p>}
          </div>
        </div>

        <div className="mt-6 flex justify-end gap-3">
          <Button ref={cancelRef} variant="outline" size="sm" onClick={onCancel} disabled={loading}>
            {cancelLabel}
          </Button>
          <Button
            variant={danger ? 'danger' : 'primary'}
            size="sm"
            onClick={onConfirm}
            disabled={loading}
          >
            {loading ? 'Procesando…' : confirmLabel}
          </Button>
        </div>
      </div>
    </div>
  )
}
