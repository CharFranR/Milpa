import { useState, useEffect, useRef } from 'react'
import Icon from '../ui/Icon'
import Button from '../ui/Button'
import { inquiries } from '../../services/api'

/**
 * @typedef {Object} ProductDetailModalProps
 * @property {boolean} isOpen
 * @property {() => void} onClose
 * @property {{id: string, name: string, producerName: string, producerPhone: string}} product
 * @property {(message: string) => void} onSend
 */

/**
 * @param {ProductDetailModalProps} props
 */
export default function ProductDetailModal({ isOpen, onClose, product, onSend }) {
  const [message, setMessage] = useState('')
  const [sending, setSending] = useState(false)
  const [error, setError] = useState('')
  const textareaRef = useRef(null)

  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden'
      setTimeout(() => textareaRef.current?.focus(), 100)
    } else {
      document.body.style.overflow = ''
    }
    return () => { document.body.style.overflow = '' }
  }, [isOpen])

  function handleSubmit(e) {
    e.preventDefault()
    if (!message.trim()) return

    setError('')
    setSending(true)

    inquiries.create({ offering_id: product.id, message: message.trim() })
      .then(() => {
        alert('Mensaje enviado correctamente.')
        setMessage('')
        onSend(message.trim())
        onClose()
      })
      .catch((err) => {
        setError(err.message || 'Error al enviar el mensaje.')
      })
      .finally(() => {
        setSending(false)
      })
  }

  function handleWhatsApp() {
    const text = encodeURIComponent(`Hola ${product.producerName}, me interesa comprar ${product.name}...`)
    const phone = product.producerPhone.replace(/\D/g, '')
    window.open(`https://wa.me/505${phone}?text=${text}`, '_blank')
    onClose()
  }

  if (!isOpen) return null

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in-200"
      role="dialog"
      aria-modal="true"
      aria-labelledby="modal-title"
    >
      <div className="relative w-full max-w-md rounded-2xl bg-white shadow-xl overflow-hidden animate-in zoom-in-95 duration-200">
        <header className="flex items-center justify-between px-5 py-4 border-b border-gray-100">
          <h3 id="modal-title" className="text-lg font-semibold text-gray-900">
            Contactar a {product.producerName}
          </h3>
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
            aria-label="Cerrar"
          >
            <Icon name="close" size={24} />
          </button>
        </header>

        <form onSubmit={handleSubmit} className="p-5 space-y-4">
          <p className="text-sm text-gray-600">
            Escribe un mensaje sobre <span className="font-medium">{product.name}</span>. El productor responderá directamente.
          </p>

          <label htmlFor="inquiry-message" className="sr-only">
            Tu mensaje
          </label>
          <textarea
            ref={textareaRef}
            id="inquiry-message"
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            placeholder={`Hola ${product.producerName}, me interesa comprar ${product.name}...`}
            rows={4}
            className="w-full rounded-xl border border-gray-200 bg-white px-4 py-3 text-sm text-gray-900 placeholder:text-gray-400 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand resize-none"
            required
            minLength={10}
            maxLength={500}
            disabled={sending}
          />

          {error && (
            <p className="rounded-lg bg-red-50 px-3 py-2 text-xs text-red-700">{error}</p>
          )}

          <div className="flex items-center justify-end gap-3 pt-2">
            <Button type="button" variant="outline" onClick={onClose} disabled={sending}>
              Cancelar
            </Button>
            <Button type="submit" variant="primary" className="whitespace-nowrap" disabled={sending}>
              {sending ? (
                <>
                  <Icon name="progress_activity" size={16} className="animate-spin" />
                  Enviando...
                </>
              ) : (
                'Enviar mensaje'
              )}
            </Button>
          </div>
        </form>

        <div className="border-t border-gray-100 p-4">
          <Button
            type="button"
            variant="whatsapp"
            className="w-full"
            icon={<Icon name="phone_iphone" size={20} />}
            onClick={handleWhatsApp}
            disabled={sending}
          >
            Contactar por WhatsApp
          </Button>
        </div>
      </div>
    </div>
  )
}