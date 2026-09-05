import Badge from '../ui/Badge'
import Button from '../ui/Button'
import ProductImage from './ProductImage'
import { formatPrice } from '../../lib/format'
import { getProductImage } from '../../lib/productImages'

export default function ProductCardList({ offering }) {
  if (!offering) return null

  const description = offering.description || ''
  const unitMatch = description.match(/Unit:\s*(\S+)/)
  const unit = unitMatch?.[1] || 'un'
  const imageUrl = offering.image_url || getProductImage(offering.id)

  return (
    <article className="group flex items-center gap-4 rounded-2xl border border-gray-100 bg-white p-3 transition-all duration-300 hover:-translate-y-0.5 hover:shadow-md motion-reduce:transition-none motion-reduce:hover:translate-y-0">
      <ProductImage
        image_url={imageUrl}
        name={offering.name}
        className="h-24 w-24 shrink-0 rounded-xl"
        imgClassName="transition-transform duration-300 group-hover:scale-105 motion-reduce:transition-none motion-reduce:group-hover:scale-100!"
      />

      <div className="min-w-0 flex-1">
        <div className="flex items-start justify-between gap-2">
          <h3 className="truncate text-base font-semibold text-gray-900">{offering.name}</h3>
          <span className="shrink-0">
            <Badge tone="brand">{offering.type === 1 ? 'Servicio' : 'Producto'}</Badge>
          </span>
        </div>
        {offering.company_name && (
          <p className="mt-0.5 truncate text-sm text-gray-500">
            {offering.company_name}
          </p>
        )}
      </div>

      <div className="flex shrink-0 flex-col items-end justify-between self-stretch py-1">
        <p className="whitespace-nowrap text-lg font-bold text-brand">
          {formatPrice(offering.price)}
          <span className="ml-1 text-sm font-medium text-gray-400">/ {unit}</span>
        </p>
        <a href={`#/product/${offering.id}`} aria-label={`Ver detalle de ${offering.name}`}>
          <Button variant="primary" size="sm">
            Ver detalle
          </Button>
        </a>
      </div>
    </article>
  )
}
