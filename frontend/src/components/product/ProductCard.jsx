import { useEffect, useState } from 'react'
import { formatPrice } from '../../lib/format'
import { categories } from '../../services/categories'
import { resolveOfferingImage } from '../../lib/productImages'
import Icon from '../ui/Icon'
import Badge from '../ui/Badge'
import ProductImage from './ProductImage'

export default function ProductCard({ offering, showArrow = true }) {
  const [categoryName, setCategoryName] = useState(offering?.category_name || null)
  const categoryId = offering?.category_id

  useEffect(() => {
    if (!categoryId || categoryName) return undefined
    let cancelled = false

    categories
      .getAll()
      .then((list) => {
        const hit = Array.isArray(list)
          ? list.find((category) => category.id === categoryId)
          : null
        if (!cancelled && hit) setCategoryName(hit.name)
      })
      .catch(() => {})

    return () => {
      cancelled = true
    }
  }, [categoryId, categoryName])

  if (!offering) return null

  const description = offering.description || ''
  const unitMatch = description.match(/Unit:\s*(\S+)/)
  const unit = unitMatch?.[1] || null
  const imageUrl = resolveOfferingImage(offering)
  const badge = categoryName || (offering.type === 1 ? 'Servicio' : 'Producto')
  const location = [offering.municipality, offering.department].filter(Boolean).join(', ')

  return (
    <article className="group relative flex flex-col overflow-hidden rounded-2xl border border-gray-100 bg-white transition-all duration-300 hover:-translate-y-1 hover:shadow-lg motion-reduce:transition-none motion-reduce:hover:translate-y-0">
      <a
        href={`#/product/${offering.id}`}
        className="relative block aspect-[4/3] overflow-hidden bg-gray-200"
        aria-hidden="true"
        tabIndex={-1}
      >
        <ProductImage
          image_url={imageUrl}
          name={offering.name}
          imgClassName="transition-transform duration-300 group-hover:scale-105 motion-reduce:transition-none motion-reduce:group-hover:scale-100!"
        />
        <span className="absolute left-3 top-3">
          <Badge tone="brand">{badge}</Badge>
        </span>
      </a>

      <div className="flex flex-1 flex-col p-4">
        <a
          href={`#/product/${offering.id}`}
          className="text-base font-semibold text-gray-900 hover:text-brand"
        >
          {offering.name}
        </a>
        {(offering.company_name || location) && (
          <p className="mt-0.5 truncate text-sm text-gray-500">
            {offering.company_name}
            {offering.company_name && location ? ' · ' : ''}
            {location}
          </p>
        )}
        <div className="mt-3 flex items-center justify-between pt-1">
          <p className="text-lg font-bold text-brand">
            {formatPrice(offering.price)}
            {unit && <span className="ml-1 text-sm font-medium text-gray-400">/ {unit}</span>}
          </p>
          {showArrow ? (
            <a
              href={`#/product/${offering.id}`}
              aria-label={`Ver detalle de ${offering.name}`}
              className="inline-flex h-9 w-9 items-center justify-center rounded-full bg-brand-soft text-brand transition-colors hover:bg-brand hover:text-white"
            >
              <Icon name="arrow_forward" size={18} />
            </a>
          ) : (
            <a
              href={`#/product/${offering.id}`}
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
