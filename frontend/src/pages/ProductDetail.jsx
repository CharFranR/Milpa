import { useState, useEffect, useRef } from 'react'
import Navbar from '../components/layout/Navbar'
import Footer from '../components/layout/Footer'
import Icon from '../components/ui/Icon'
import Button from '../components/ui/Button'
import Badge from '../components/ui/Badge'
import ProductImage from '../components/product/ProductImage'
import { offerings, companies, conversations, openChat } from '../services/api'
import { productById, producerById, categoryById } from '../mocks/catalog'
import { isAuthenticated } from '../lib/session'
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
  const [chatError, setChatError] = useState('')
  const [startingChat, setStartingChat] = useState(false)
  const [chat, setChat] = useState({ conversationId: null, messages: [], connected: false })
  const [draft, setDraft] = useState('')
  const socketRef = useRef(null)

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

  useEffect(() => {
    return () => {
      socketRef.current?.close()
      socketRef.current = null
    }
  }, [])

  const product = realOffering || productById(productId)
  // Only the public representation is available to an anonymous visitor, so
  // this is the location a buyer can be shown and nothing finer.
  const producer = realCompany
    ? {
        name: realCompany.name,
        city: realCompany.municipality || 'Nicaragua',
        region: realCompany.department || '',
        farm: realCompany.description || '',
        verified: !!realCompany.verified,
        since: '2025',
      }
    : product
      ? { ...producerById(product.producerId), verified: false }
      : null
  const category = product ? (categoryById(product.categoryId) || { name: product.type === 1 ? 'Servicio' : 'Producto' }) : null
  const farmerId = realOffering?.user_id || null

  const description = product?.description || ''
  const unitMatch = description.match(/Unit:\s*(\S+)/)
  const unit = unitMatch?.[1] || 'un'
  const qtyMatch = description.match(/Qty:\s*(\d+)/)
  const quantity = qtyMatch?.[1] || null
  const cleanDescription = description.replace(/Unit:\s*\S+\n?/, '').replace(/Qty:\s*\d+\n?/, '').replace(/Category:\s*.+\n?/, '').trim()

  // The direct chat is the only contact path on this page. Post-match chat does
  // not exist yet, so this opens a conversation immediately instead of waiting
  // for a match; that is temporary and is the reason no phone number is shown.
  async function handleContact() {
    setChatError('')

    if (!isAuthenticated()) {
      window.location.hash = '#/login'
      return
    }
    if (!farmerId) {
      setChatError('Este anuncio no tiene un productor registrado todavía.')
      return
    }

    setStartingChat(true)
    try {
      const conversation = await conversations.create({
        farmer_id: farmerId,
        offering_id: product.id,
      })
      setIsModalOpen(true)
      setChat({ conversationId: conversation.id, messages: [], connected: false })

      socketRef.current?.close()
      const socket = openChat(conversation.id, {
        onOpen: () => setChat((prev) => ({ ...prev, connected: true })),
        onMessage: (message) => {
          setChat((prev) => ({ ...prev, messages: [...prev.messages, message] }))
        },
        onClose: () => setChat((prev) => ({ ...prev, connected: false })),
        onError: () => setChatError('No se pudo conectar el chat.'),
      })
      socketRef.current = socket
    } catch (err) {
      setChatError(err?.message || 'No se pudo iniciar la conversación.')
    } finally {
      setStartingChat(false)
    }
  }

  function handleSend(event) {
    event.preventDefault()
    const text = draft.trim()
    if (!text) return
    socketRef.current?.send(text)
    setDraft('')
  }

  function handleSendMessage() {
    setIsModalOpen(false)
    socketRef.current?.close()
    socketRef.current = null
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
                disabled={startingChat}
                onClick={handleContact}
              >
                {startingChat ? 'Abriendo chat...' : 'Contactar productor'}
              </Button>

              {chatError && (
                <p className="rounded-lg bg-red-50 px-3 py-2 text-xs text-red-700">{chatError}</p>
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
                  {producer.verified && (
                    <span className="mt-1 inline-flex items-center gap-1 rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700">
                      <Icon name="check_circle" size={12} />
                      Productor verificado
                    </span>
                  )}
                </div>
              </div>
            </div>
          </section>
        </div>
      </main>

      <Footer />

      {isModalOpen && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in-200"
          role="dialog"
          aria-modal="true"
          aria-label={`Chat con ${producer.name}`}
        >
          <div className="relative w-full max-w-md rounded-2xl bg-white shadow-xl overflow-hidden">
            <header className="flex items-center justify-between px-5 py-4 border-b border-gray-100">
              <h3 className="text-lg font-semibold text-gray-900">{producer.name}</h3>
              <button
                type="button"
                onClick={handleSendMessage}
                className="rounded-lg p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
                aria-label="Cerrar"
              >
                <Icon name="close" size={24} />
              </button>
            </header>

            <div className="max-h-80 space-y-2 overflow-y-auto p-5">
              {chat.messages.length === 0 && (
                <p className="text-sm text-gray-500">
                  Escribe tu consulta sobre {product.name}. El productor responderá por aquí.
                </p>
              )}
              {chat.messages.map((message, index) => (
                <p
                  key={`${message.id || 'msg'}-${index}`}
                  className="rounded-xl bg-gray-50 px-3 py-2 text-sm text-gray-800"
                >
                  {message.content}
                </p>
              ))}
            </div>

            <form onSubmit={handleSend} className="flex items-center gap-2 border-t border-gray-100 p-4">
              <label htmlFor="chat-message" className="sr-only">Tu mensaje</label>
              <input
                id="chat-message"
                value={draft}
                onChange={(e) => setDraft(e.target.value)}
                placeholder="Escribe un mensaje..."
                className="flex-1 rounded-xl border border-gray-200 px-4 py-2 text-sm text-gray-900 placeholder:text-gray-400 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
                disabled={!chat.connected}
                maxLength={2000}
              />
              <Button type="submit" variant="primary" disabled={!chat.connected || !draft.trim()}>
                Enviar
              </Button>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
