import { useState } from 'react'
import { cn } from '../../lib/cn'
import { productById, producerById, categoryById } from '../../mocks/catalog'
import { formatPrice } from '../../lib/format'
import Icon from '../ui/Icon'
import Badge from '../ui/Badge'
import ProductImage from './ProductImage'
import StarRating from '../StarRating'

export default function ProductCard({
  productId,
  initiallyFavorite = false,
  onToggleFavorite,
  showArrow = false,
}) {
  const product = productById(productId)
  const producer = producerById(product.producerId)
  const category = categoryById(product.categoryId)
  const [favorite, setFavorite] = useState(initiallyFavorite)

  if (!product || !producer) return null

  function toggleFavorite() {
    const next = !favorite
    setFavorite(next)
    onToggleFavorite?.(product.id, next)
  }

  return (
    <article className="group relative flex flex-col overflow-hidden rounded-2xl border border-gray-100 bg-white transition-all duration-300 hover:-translate-y-1 hover:shadow-lg">
      <a
        href={`#/product/${product.id}`}
        className="relative block aspect-[4/3] overflow-hidden bg-gray-200"
        aria-hidden="true"
        tabIndex={-1}
      >
        <ProductImage
          productId={product.id}
          name={product.name}
          imgClassName="transition-transform duration-300 group-hover:scale-105"
        />
        <span className="absolute left-3 top-3">
          <Badge tone="brand">{category.name}</Badge>
        </span>
      </a>

      <button
        type="button"
        onClick={toggleFavorite}
        aria-pressed={favorite}
        aria-label={favorite ? 'Quitar de favoritos' : 'Agregar a favoritos'}
        className={cn(
          'absolute right-3 top-3 inline-flex h-9 w-9 items-center justify-center rounded-full bg-white/90 shadow-sm transition-colors hover:bg-white',
          favorite ? 'text-red-500' : 'text-gray-400 hover:text-red-500',
        )}
      >
        <Icon name="favorite" size={18} filled={favorite} />
      </button>

      <div className="flex flex-1 flex-col p-4">
        <a href={`#/product/${product.id}`} className="text-base font-semibold text-gray-900 hover:text-brand">
          {product.name}
        </a>
        <p className="mt-0.5 text-sm text-gray-500">
          {producer.name} · {producer.city}
        </p>
        <div className="mt-2 flex items-center gap-2">
          <StarRating rating={product.rating} reviews={product.reviews} size={14} />
        </div>
        <div className="mt-3 flex items-center justify-between pt-1">
          <p className="text-lg font-bold text-brand">
            {formatPrice(product.price)}
            <span className="ml-1 text-sm font-medium text-gray-400">/ {product.unit}</span>
          </p>
          {showArrow ? (
            <a
              href={`#/product/${product.id}`}
              aria-label={`Ver detalle de ${product.name}`}
              className="inline-flex h-9 w-9 items-center justify-center rounded-full bg-brand-soft text-brand transition-colors hover:bg-brand hover:text-white"
            >
              <Icon name="arrow_forward" size={18} />
            </a>
          ) : (
            <a
              href={`#/product/${product.id}`}
              className="rounded-full px-4 py-2 text-sm font-semibold text-brand transition-colors hover:bg-brand-soft"
            >
              Ver detalle
            </a>
          )}
        </div>
      </div>
    </article>
  )
}