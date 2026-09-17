import { useState, useEffect } from 'react'
import Navbar from '../components/layout/Navbar'
import Footer from '../components/layout/Footer'
import Icon from '../components/ui/Icon'
import Button from '../components/ui/Button'
import Badge from '../components/ui/Badge'
import ProductImage from '../components/product/ProductImage'
import ProductDetailModal from '../components/product/ProductDetailModal'
import { offerings, companies } from '../services/api'
import { productById, producerById, categoryById } from '../mocks/catalog'
import { formatPrice } from '../lib/format'
import { cn } from '../lib/cn'
import { resolveOfferingImage } from '../lib/productImages'

export default function ProductDetail() {
  const hash = window.location.hash
  const productId = hash.replace('#/product/', '')
  const [realOffering, setRealOffering] = useState(null)
  const [realCompany, setRealCompany] = useState(null)
  const [loading, setLoading] = useState(true)
  const [isModalOpen, setIsModalOpen] = useState(false)

  useEffect(() => {
    if (!productId) return
    setLoading(true)

    offerings.getById(productId)
      .then((data) => {
        setRealOffering(data)
        if (data.company_id) {
          return companies.getById(data.company_id).catch(() => null)
        }
        return null
      })
      .then((companyData) => {
        if (companyData) setRealCompany(companyData)
      })
      .catch(() => {
        setRealOffering(null)
      })
      .finally(() => setLoading(false))
  }, [productId])

  const product = realOffering || productById(productId)
  const producer = realCompany
    ? {
        name: realCompany.name,
        city: realCompany.address || 'Nicaragua',
        region: realCompany.address || '',
        phone: realCompany.phone_number || '',
        farm: realCompany.description || '',
        since: '2025',
      }
    : product ? producerById(product.producerId) : null
  const category = product ? (categoryById(product.categoryId) || { name: product.type === 1 ? 'Servicio' : 'Producto' }) : null

  const description = product?.description || ''
  const unitMatch = description.match(/Unit:\s*(\S+)/)
  const unit = unitMatch?.[1] || 'un'
  const qtyMatch = description.match(/Qty:\s*(\d+)/)
  const quantity = qtyMatch?.[1] || null
  const cleanDescription = description.replace(/Unit:\s*\S+\n?/, '').replace(/Qty:\s*\d+\n?/, '').replace(/Category:\s*.+\n?/, '').trim()

  function handleSendMessage(message) {
    console.log('Mensaje enviado:', message)
    setIsModalOpen(false)
  }

  if (loading) {
    return (
      <div className="flex min-h-screen flex-col bg-gray-50">
        <Navbar />
        <main className="mx-auto flex-1 flex items-center justify-center px-4 py-16">
          <div className="text-center">
            <div className="h-12 w-12 mx-auto rounded-full bg-brand-soft animate-pulse" />
            <p className="mt-4 text-sm text-gray-500">Cargando producto...</p>
          </div>
        </main>
        <Footer />
      </div>
    )
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

  const availabilityText = realOffering ? 'Disponible ahora' : 'Disponible ahora'
  const availabilityColor = 'bg-green-100 text-green-700'

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
                image_url={resolveOfferingImage(realOffering)}
                name={product.name}
                className="w-full h-full object-cover"
              />
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

            <div className="text-3xl font-bold text-brand">
              {formatPrice(product.price)}
              <span className="text-base font-medium text-gray-400"> / {unit}</span>
            </div>

            <div className="flex flex-wrap items-center gap-2">
              <span className={cn('inline-flex items-center gap-1 rounded-full px-3 py-1 text-xs font-medium', availabilityColor)}>
                <Icon name="check_circle" size={12} />
                {availabilityText}
              </span>
              {producer.region && (
                <span className="inline-flex items-center gap-1 rounded-full bg-gray-100 px-3 py-1 text-xs font-medium text-gray-700">
                  <Icon name="location_on" size={12} />
                  {producer.region}
                </span>
              )}
            </div>

            <div className="border-t border-gray-100 pt-6 space-y-2">
              {cleanDescription && (
                <p className="text-sm text-gray-600 leading-relaxed">{cleanDescription}</p>
              )}
              {quantity && (
                <p className="text-sm text-gray-500">Cantidad disponible: {quantity}</p>
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

              {producer.phone && (
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
              )}
            </div>

            <div className="rounded-xl border border-gray-100 bg-gray-50 p-5 space-y-4">
              <div className="flex items-center gap-3">
                <span className="flex h-14 w-14 items-center justify-center rounded-full bg-brand-soft text-brand">
                  <Icon name="agriculture" size={24} />
                </span>
                <div>
                  <p className="font-semibold text-gray-900">{producer.name}</p>
                  {producer.farm && (
                    <p className="text-sm text-gray-500">{producer.farm}</p>
                  )}
                  <p className="text-xs text-gray-500 flex items-center gap-1">
                    <Icon name="location_on" size={12} />
                    {producer.city} · Miembro desde {producer.since}
                  </p>
                </div>
              </div>
            </div>
          </section>
        </div>
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
