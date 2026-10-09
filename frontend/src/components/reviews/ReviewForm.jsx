import { useState } from 'react'
import Button from '../ui/Button'
import Icon from '../ui/Icon'
import { cn } from '../../lib/cn'
import { friendlyReviewError } from '../../lib/reviewMessages'

const STARS = [1, 2, 3, 4, 5]

export default function ReviewForm({ targetLabel = 'a la otra parte', onSubmit, disabled = false }) {
  const [rating, setRating] = useState(0)
  const [hovered, setHovered] = useState(0)
  const [comment, setComment] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  const visible = hovered || rating

  async function handleSubmit(event) {
    event.preventDefault()
    setError('')

    if (rating < 1 || rating > 5) {
      setError('Selecciona una valoración de 1 a 5 estrellas.')
      return
    }

    setBusy(true)
    try {
      await onSubmit({ rating, comment: comment.trim() })
      setRating(0)
      setComment('')
    } catch (err) {
      setError(friendlyReviewError(err))
    } finally {
      setBusy(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <span className="text-xs font-semibold uppercase tracking-wide text-gray-400">
          Tu valoración
        </span>
        <div className="mt-2 flex items-center gap-1" role="radiogroup" aria-label="Valoración">
          {STARS.map((value) => (
            <button
              key={value}
              type="button"
              role="radio"
              aria-checked={rating === value}
              aria-label={`${value} estrella${value > 1 ? 's' : ''}`}
              disabled={busy || disabled}
              onClick={() => setRating(value)}
              onMouseEnter={() => setHovered(value)}
              onMouseLeave={() => setHovered(0)}
              className="rounded p-1 text-amber-400 transition-transform hover:scale-110 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand disabled:cursor-not-allowed"
            >
              <Icon
                name="star"
                size={28}
                filled={value <= visible}
                weight={value <= visible ? 600 : 400}
                className={value <= visible ? '' : 'text-gray-300'}
              />
            </button>
          ))}
          <span className="ml-2 text-sm font-medium text-gray-500">
            {visible > 0 ? `${visible}/5` : 'Sin valorar'}
          </span>
        </div>
      </div>

      <div>
        <label htmlFor="review_comment" className="block text-xs font-semibold text-gray-600">
          Comentario (opcional)
        </label>
        <textarea
          id="review_comment"
          rows={3}
          value={comment}
          onChange={(event) => setComment(event.target.value)}
          placeholder="¿Cómo te fue con la transacción?"
          disabled={busy || disabled}
          className="mt-1.5 w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2.5 text-sm text-gray-900 placeholder:text-gray-400 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand disabled:opacity-60"
        />
      </div>

      {error && (
        <p role="alert" className="text-sm text-red-600">
          {error}
        </p>
      )}

      <div className={cn('flex flex-wrap items-center gap-3')}>
        <Button type="submit" disabled={busy || disabled}>
          {busy ? 'Enviando…' : `Enviar reseña ${targetLabel}`}
        </Button>
        <p className="text-xs text-gray-500">Solo puedes dejar una reseña por transacción.</p>
      </div>
    </form>
  )
}
