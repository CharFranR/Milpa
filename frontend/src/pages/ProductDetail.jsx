import { useState, useEffect } from 'react'
import Navbar from '../components/layout/Navbar'
import Footer from '../components/layout/Footer'
import Icon from '../components/ui/Icon'
import Button from '../components/ui/Button'
import Badge from '../components/ui/Badge'
import StarRating from '../components/StarRating'
import ProductImage from '../components/product/ProductImage'
import ProductDetailModal from '../components/product/ProductDetailModal'
import { productById, producerById, categoryById, productsByCategory } from '../mocks/catalog'
import { formatPrice } from '../lib/format'
import { cn } from '../lib/cn'

export default function ProductDetail() {
  const hash = window.location.hash
  const productId = hash.replace('#/product/', '')
  const product = productById(productId)
  const producer = product ? producerById(product.producerId) : null
  const category = product ? categoryById(product.categoryId) : null

  const [isModalOpen, setIsModalOpen] = useState(false)
  const [relatedProducts, setRelatedProducts] = useState([])

  useEffect(() => {
    if (product) {
      const related = productsByCategory(product.categoryId)
        .filter((p) => p.id !== product.id)
        .slice(0, 3)
      setRelatedProducts(related)
    }
  }, [product])

function handleSendMessage(message) {
  console.log('Mensaje enviado:', message)
  setIsModalOpen(false)
}

  if (!product || !producer || !category) {
    return (
      <div className="flex min-h-screen flex-col bg-gray-50">
        <Navbar />
        <main className="mx-auto flex-1 flex items-center justify-center px-4 py-16">
          <div className="text-center">
            <Icon name="error" size={48} className="mx-auto text-gray-400" />
            <h1 className="mt-4 text-xl font-bold text-gray-900">Producto no encontrado</h1>
            <p className="mt-2 text-gray-500">El producto que buscas no existe o ha sido eliminado.</p>
            <a href="#/marketplace" className="mt-6 inline-block text-brand hover:underline">
              Volver al Marketplace
            </a>
          </div>
        </main>
        <Footer />
      </div>
    )
  }

  const availabilityText = product.available ? 'Disponible ahora' : 'Temporalmente agotado'
  const availabilityColor = product.available ? 'bg-green-100 text-green-700' : 'bg-yellow-100 text-yellow-700'

  return (
    <div className="flex min-h-screen flex-col bg-gray-50">
      <Navbar />

      <main className="mx-auto flex-1 px-4 py-8 sm:px-6 lg:px-8">
        <nav aria-label="Ruta de navegación" className="mb-6 flex items-center gap-1 text-sm text-gray-500">
          <a href="#/" className="hover:text-brand">Inicio</a>
          <Icon name="chevron_right" size={16} className="text-gray-300" />
          <a href="#/marketplace" className="hover:text-brand">Marketplace</a>
          <Icon name="chevron_right" size={16} className="text-gray-300" />
          <span className="font-semibold text-gray-900 truncate max-w-xs">{product.name}</span>
        </nav>

        <div className="grid gap-8 lg:grid-cols-3">
          <section aria-label="Galería de imágenes" className="lg:col-span-2 space-y-4">
            <div className="relative aspect-[4/3] rounded-2xl overflow-hidden bg-gray-100">
              <ProductImage
                productId={product.id}
                name={product.name}
                className="w-full h-full object-cover"
              />
            </div>
            <div className="flex gap-2 overflow-x-auto pb-2">
              {Array.from({ length: 4 }, (_, i) => (
                <button
                  key={i}
                  type="button"
                  className={cn(
                    'flex-shrink-0 w-20 h-20 rounded-xl overflow-hidden border-2 transition-colors',
                    i === 0 ? 'border-brand' : 'border-transparent hover:border-brand/40',
                  )}
                  aria-label={`Miniatura ${i + 1}`}
                >
                  <ProductImage
                    productId={product.id}
                    name={product.name}
                    className="w-full h-full object-cover"
                  />
                </button>
              ))}
            </div>
          </section>

          <section aria-label="Información del producto" className="space-y-6">
            <div className="flex items-start justify-between gap-2">
              <Badge tone="brand">{category.name}</Badge>
              <button
                type="button"
                className="shrink-0 rounded-xl p-2 text-gray-400 hover:text-red-500 hover:bg-red-50"
                aria-label="Agregar a favoritos"
              >
                <Icon name="favorite" size={22} />
              </button>
            </div>

            <h1 className="text-2xl font-bold text-gray-900 sm:text-3xl">{product.name}</h1>

            <div className="flex items-center gap-2">
              <StarRating rating={product.rating} reviews={product.reviews} size={16} showValue />
            </div>

            <div className="text-3xl font-bold text-brand">
              {formatPrice(product.price)}
              <span className="text-base font-medium text-gray-400"> / {product.unit}</span>
            </div>

            <div className="flex flex-wrap items-center gap-2">
              <span className={cn('inline-flex items-center gap-1 rounded-full px-3 py-1 text-xs font-medium', availabilityColor)}>
                <Icon name={product.available ? 'check_circle' : 'schedule'} size={12} />
                {availabilityText}
              </span>
              <span className="inline-flex items-center gap-1 rounded-full bg-gray-100 px-3 py-1 text-xs font-medium text-gray-700">
                <Icon name="location_on" size={12} />
                {producer.region}
              </span>
            </div>

            <div className="border-t border-gray-100 pt-6 space-y-2">
              <p className="text-sm text-gray-600 leading-relaxed">{product.description}</p>
              {product.tags.length > 0 && (
                <div className="flex flex-wrap gap-1.5">
                  {product.tags.map((tag) => (
                    <span key={tag} className="inline-flex items-center gap-1 rounded-full bg-brand-soft px-2.5 py-0.5 text-xs font-medium text-brand">
                      #{tag}
                    </span>
                  ))}
                </div>
              )}
            </div>

            <div className="border-t border-gray-100 pt-6 space-y-3">
              <Button
                type="button"
                variant="primary"
                size="lg"
                className="w-full"
                icon={<Icon name="chat_bubble" size={20} />}
                onClick={() => setIsModalOpen(true)}
              >
                Contactar productor
              </Button>

              <Button
                type="button"
                variant="whatsapp"
                size="lg"
                className="w-full"
                icon={<Icon name="phone_iphone" size={20} />}
                onClick={() => {
                  const text = encodeURIComponent(`Hola ${producer.name}, me interesa comprar ${product.name}...`)
                  const phone = producer.phone.replace(/\D/g, '')
                  window.open(`https://wa.me/505${phone}?text=${text}`, '_blank')
                }}
              >
                Contactar por WhatsApp
              </Button>
            </div>

            <div className="rounded-xl border border-gray-100 bg-gray-50 p-5 space-y-4">
              <div className="flex items-center gap-3">
                <span className="flex h-14 w-14 items-center justify-center rounded-full bg-brand-soft text-brand">
                  <Icon name="agriculture" size={24} />
                </span>
                <div>
                  <p className="font-semibold text-gray-900">{producer.name}</p>
                  <p className="text-sm text-gray-500">{producer.farm}</p>
                  <p className="text-xs text-gray-500 flex items-center gap-1">
                    <Icon name="location_on" size={12} />
                    {producer.city}, {producer.region} · Miembro desde {producer.since}
                  </p>
                </div>
              </div>
              <a href="#/marketplace" className="inline-flex items-center gap-1 text-sm font-semibold text-brand hover:text-brand-dark">
                Ver perfil <Icon name="arrow_forward" size={14} />
              </a>
            </div>
          </section>
        </div>

        {relatedProducts.length > 0 && (
          <section aria-label="Productos relacionados" className="mt-12">
            <h2 className="text-xl font-bold text-gray-900 mb-6">Productos relacionados</h2>
            <div className="grid gap-5 sm:grid-cols-3">
              {relatedProducts.map((related) => {
                const relatedProducer = producerById(related.producerId)
                const relatedCategory = categoryById(related.categoryId)
                return (
                  <article
                    key={related.id}
                    className="group overflow-hidden rounded-2xl border border-gray-100 bg-white transition-all duration-300 hover:-translate-y-1 hover:shadow-lg"
                  >
                    <a
                      href={`#/product/${related.id}`}
                      className="relative block aspect-[4/3] overflow-hidden bg-gray-100"
                      aria-label={`Ver ${related.name}`}
                    >
                      <ProductImage
                        productId={related.id}
                        name={related.name}
                        className="transition-transform duration-300 group-hover:scale-105"
                      />
                      <span className="absolute left-3 top-3">
                        <Badge tone="brand">{relatedCategory?.name}</Badge>
                      </span>
                    </a>
                    <div className="p-4">
                      <a href={`#/product/${related.id}`} className="font-semibold text-gray-900 hover:text-brand line-clamp-1">
                        {related.name}
                      </a>
                      <p className="mt-0.5 text-sm text-gray-500 truncate">
                        {relatedProducer?.name} · {relatedProducer?.city}
                      </p>
                      <div className="mt-2 flex items-center justify-between">
                        <p className="text-lg font-bold text-brand">
                          {formatPrice(related.price)}
                          <span className="ml-1 text-sm font-medium text-gray-400">/ {related.unit}</span>
                        </p>
                      </div>
                    </div>
                  </article>
                )
              })}
            </div>
          </section>
        )}

      </main>

      <Footer />

      <ProductDetailModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        product={{
          id: product.id,
          name: product.name,
          producerName: producer.name,
          producerPhone: producer.phone,
        }}
        onSend={handleSendMessage}
      />
    </div>
  )
}