import SectionHeading from '../../components/SectionHeading'
import ProductCard from '../../components/product/ProductCard'
import { products } from '../../mocks/catalog'

export default function FeaturedProducts() {
  const featured = products.slice(0, 4)

  return (
    <section className="bg-gray-50">
      <div className="mx-auto max-w-7xl px-4 py-16 sm:px-6 lg:px-8">
        <SectionHeading
          eyebrow="Lo más fresco hoy"
          title="Productos destacados"
          action={
            <a href="#" className="group inline-flex items-center gap-1 text-sm font-semibold text-brand hover:text-brand-dark">
              Ver todo el catálogo
              <span aria-hidden="true" className="transition-transform group-hover:translate-x-0.5">→</span>
            </a>
          }
        />
        <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
          {featured.map((product) => (
            <ProductCard key={product.id} productId={product.id} />
          ))}
        </div>
      </div>
    </section>
  )
}