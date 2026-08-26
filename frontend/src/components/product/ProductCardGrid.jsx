import Badge from '../ui/Badge'
import Icon from '../ui/Icon'
import ProductImage from './ProductImage'
import StarRating from '../StarRating'
import { formatPrice } from '../../lib/format'
import { categoryById, producerById, productById } from '../../mocks/catalog'

export default function ProductCardGrid({ productId }) {
  const product = productById(productId)
  if (!product) return null

  const producer = producerById(product.producerId)
  const category = categoryById(product.categoryId)

  return (
    <article className="group relative flex flex-col overflow-hidden rounded-2xl border border-gray-100 bg-white transition-all duration-300 hover:-translate-y-1 hover:shadow-lg motion-reduce:transition-none motion-reduce:hover:translate-y-0">
      <div className="relative aspect-[4/3]">
        <ProductImage
          productId={product.id}
          name={product.name}
          imgClassName="transition-transform duration-300 group-hover:scale-105 motion-reduce:transition-none motion-reduce:group-hover:!scale-100"
        />
        <span className="absolute left-3 top-3">
          <Badge tone="brand">{category.name}</Badge>
        </span>
      </div>

      <div className="flex flex-1 flex-col p-4">
        <h3 className="text-base font-semibold text-gray-900">{product.name}</h3>
        <p className="mt-0.5 text-sm text-gray-500">
          {producer.name} · {producer.city}
        </p>
        <StarRating rating={product.rating} size={14} showValue className="mt-2" />

        <div className="mt-3 flex items-center justify-between pt-1">
          <p className="text-lg font-bold text-brand">
            {formatPrice(product.price)}
            <span className="ml-1 text-sm font-medium text-gray-400">/ {product.unit}</span>
          </p>
          <a
            href="#"
            aria-label={`Ver detalle de ${product.name}`}
            className="inline-flex h-9 w-9 items-center justify-center rounded-full bg-brand-soft text-brand transition-colors hover:bg-brand hover:text-white"
          >
            <Icon name="arrow_forward" size={18} />
          </a>
        </div>
      </div>
    </article>
  )
}
