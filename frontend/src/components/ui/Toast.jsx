import { useEffect } from 'react'
import { cn } from '../../lib/cn'
import Icon from './Icon'

const TONES = {
  success: 'bg-night text-white',
  info: 'bg-brand text-white',
  error: 'bg-red-600 text-white',
}

export default function Toast({ message, tone = 'success', duration = 3000, onClose, className }) {
  useEffect(() => {
    if (!message || !onClose || !duration) return undefined
    const timer = setTimeout(onClose, duration)
    return () => clearTimeout(timer)
  }, [message, onClose, duration])

  if (!message) return null

  return (
    <div
      role="status"
      className={cn(
        'fixed bottom-6 left-1/2 z-50 flex -translate-x-1/2 items-center gap-3 rounded-full px-5 py-3 text-sm font-medium shadow-lg',
        TONES[tone],
        className,
      )}
    >
      <span>{message}</span>
      {onClose && (
        <button
          type="button"
          onClick={onClose}
          aria-label="Cerrar aviso"
          className="rounded-full p-0.5 opacity-70 transition-opacity hover:opacity-100"
        >
          <Icon name="close" size={16} />
        </button>
      )}
    </div>
  )
}
