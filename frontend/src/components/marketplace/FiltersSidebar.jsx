import { useState, useEffect } from 'react'
import Icon from '../ui/Icon'
import StarRating from '../StarRating'
import CategoryPill from '../CategoryPill'
import { formatPrice } from '../../lib/format'
import { categories } from '../../services/api'
import { cn } from '../../lib/cn'

export const PRICE_LIMIT = 5000

const MIN_RATING_OPTIONS = [
  { value: 0, label: 'Todas' },
  { value: 3.5, label: '3.5+' },
  { value: 4.0, label: '4.0+' },
  { value: 4.5, label: '4.5+' },
]

export default function FiltersSidebar({ filters, onChange, onClear }) {
  const [cats, setCats] = useState([])

  useEffect(() => {
    categories.getAll()
      .then((data) => setCats(Array.isArray(data) ? data : []))
      .catch(() => {})
  }, [])

  return (
    <div className="rounded-2xl border border-gray-100 bg-white p-5">
      <div className="flex items-center justify-between">
        <h2 className="text-base font-bold text-gray-900">Filtros</h2>
        <button
          type="button"
          onClick={onClear}
          className="text-sm font-semibold text-brand hover:text-brand-dark"
        >
          Limpiar
        </button>
      </div>

      <section aria-label="Filtrar por categoría" className="mt-5">
        <h3 className="text-xs font-semibold uppercase tracking-wider text-gray-400">Categoría</h3>
        <div className="mt-3 space-y-2">
          <button
            type="button"
            onClick={() => onChange({ category: 'all' })}
            aria-pressed={filters.category === 'all'}
            className={cn(
              'flex w-full items-center gap-2.5 rounded-xl border px-4 py-2.5 text-sm font-medium transition-colors',
              'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand',
              filters.category === 'all'
                ? 'border-brand bg-brand text-white'
                : 'border-gray-200 bg-white text-gray-700 hover:border-brand/40 hover:bg-brand-soft',
            )}
          >
            <Icon name="apps" size={18} className={filters.category === 'all' ? '' : 'text-brand'} />
            Todas
          </button>
          {cats.map((category) => (
            <CategoryPill
              key={category.id}
              icon={category.icon || 'label'}
              name={category.name}
              count={category.count}
              active={filters.category === category.id}
              onClick={() => onChange({ category: filters.category === category.id ? 'all' : category.id })}
            />
          ))}
        </div>
      </section>

      <section aria-label="Precio máximo por unidad" className="mt-6 border-t border-gray-100 pt-5">
        <h3 className="text-xs font-semibold uppercase tracking-wider text-gray-400">
          Precio por unidad
        </h3>
        <p className="mt-3 text-sm font-semibold text-gray-700">
          {formatPrice(0)} – {formatPrice(filters.maxPrice)}
        </p>
        <label htmlFor="price-range" className="sr-only">
          Precio máximo por unidad
        </label>
        <input
          id="price-range"
          type="range"
          min={0}
          max={PRICE_LIMIT}
          step={500}
          value={filters.maxPrice}
          onChange={(e) => onChange({ maxPrice: Number(e.target.value) })}
          className="mt-2 w-full accent-brand"
        />
      </section>

      <section aria-label="Valoración mínima" className="mt-6 border-t border-gray-100 pt-5">
        <h3 className="text-xs font-semibold uppercase tracking-wider text-gray-400">
          Valoración mínima
        </h3>
        <div className="mt-3 space-y-2">
          {MIN_RATING_OPTIONS.map((option) => (
            <button
              key={option.value}
              type="button"
              onClick={() => onChange({ minRating: option.value })}
              aria-pressed={filters.minRating === option.value}
              className={cn(
                'flex w-full items-center justify-between rounded-xl border px-4 py-2.5 text-sm font-medium transition-colors',
                'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand',
                filters.minRating === option.value
                  ? 'border-brand bg-brand-soft text-brand'
                  : 'border-gray-200 bg-white text-gray-700 hover:border-brand/40',
              )}
            >
              {option.label}
              {option.value > 0 && <StarRating rating={option.value} size={14} />}
            </button>
          ))}
        </div>
      </section>
    </div>
  )
}
