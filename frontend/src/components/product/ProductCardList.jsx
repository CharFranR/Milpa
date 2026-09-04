import Badge from '../ui/Badge'
import Button from '../ui/Button'
import ProductImage from './ProductImage'
import StarRating from '../StarRating'
import { formatPrice } from '../../lib/format'
import { categoryById, producerById, productById } from '../../mocks/catalog'

export default function ProductCardList({ productId }) {
  const product = productById(productId)
  if (!product) return null

  const producer = producerById(product.producerId)
  const category = categoryById(product.categoryId)

  return (
    <article className="group flex items-center gap-4 rounded-2xl border border-gray-100 bg-white p-3 transition-all duration-300 hover:-translate-y-0.5 hover:shadow-md motion-reduce:transition-none motion-reduce:hover:translate-y-0">
      <ProductImage
        productId={product.id}
        name={product.name}
        className="h-24 w-24 shrink-0 rounded-xl"
        imgClassName="transition-transform duration-300 group-hover:scale-105 motion-reduce:transition-none motion-reduce:group-hover:scale-100!"
      />

      <div className="min-w-0 flex-1">
        <div className="flex items-start justify-between gap-2">
          <h3 className="truncate text-base font-semibold text-gray-900">{product.name}</h3>
          <span className="shrink-0">
            <Badge tone="brand">{category.name}</Badge>
          </span>
        </div>
        <p className="mt-0.5 truncate text-sm text-gray-500">
          {producer.name} · {producer.city}
        </p>
        <StarRating rating={product.rating} reviews={product.reviews} size={14} className="mt-1.5" />
      </div>

      <div className="flex shrink-0 flex-col items-end justify-between self-stretch py-1">
        <p className="whitespace-nowrap text-lg font-bold text-brand">
          {formatPrice(product.price)}
          <span className="ml-1 text-sm font-medium text-gray-400">/ {product.unit}</span>
        </p>
        <a href={`#/product/${product.id}`} aria-label={`Ver detalle de ${product.name}`}>
          <Button variant="primary" size="sm">
            Ver detalle
          </Button>
        </a>
      </div>
    </article>
  )
}
