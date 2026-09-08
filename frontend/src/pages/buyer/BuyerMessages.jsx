import { useState } from 'react'
import Icon from '../../components/ui/Icon'
import Avatar from '../../components/Avatar'
import { cn } from '../../lib/cn'
import { activeChatMessages, conversations } from '../../mocks/buyer'

export default function BuyerMessages() {
  const [activeId, setActiveId] = useState(conversations[0].id)
  const [messages, setMessages] = useState(activeChatMessages)
  const [draft, setDraft] = useState('')

  const activeConversation = conversations.find((conversation) => conversation.id === activeId)

  function handleSend(e) {
    e.preventDefault()
    if (!draft.trim()) return
    setMessages((current) => [...current, { from: 'me', text: draft.trim() }])
    setDraft('')
  }

  return (
    <div className="space-y-6">
      <header>
        <h1 className="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">Mensajes</h1>
        <p className="mt-1 text-sm text-gray-500">Coordina tus compras directamente con productores.</p>
      </header>

      <div className="grid gap-4 lg:grid-cols-[minmax(280px,340px)_1fr]">
        <ul
          role="list"
          aria-label="Conversaciones"
          className="h-fit divide-y divide-gray-100 overflow-hidden rounded-2xl border border-gray-100 bg-white"
        >
          {conversations.map((conversation) => (
            <li key={conversation.id}>
              <button
                type="button"
                onClick={() => setActiveId(conversation.id)}
                aria-current={activeId === conversation.id || undefined}
                className={cn(
                  'flex w-full items-center gap-3 px-4 py-4 text-left transition-colors',
                  activeId === conversation.id ? 'bg-brand-soft' : 'hover:bg-gray-50',
                )}
              >
                <Avatar initials={conversation.initials} name={conversation.name} />
                <span className="min-w-0 flex-1">
                  <span className="block truncate text-sm font-bold text-gray-900">{conversation.name}</span>
                  <span className="block truncate text-xs text-gray-500">{conversation.lastMessage}</span>
                </span>
                <span className="flex shrink-0 flex-col items-end gap-1">
                  <time className="text-[11px] text-gray-400">{conversation.time}</time>
                  {conversation.unread > 0 && (
                    <span className="inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-brand px-1.5 text-[11px] font-bold text-white">
                      {conversation.unread}
                    </span>
                  )}
                </span>
              </button>
            </li>
          ))}
        </ul>

        <section
          aria-label={`Conversación activa con ${activeConversation.name}`}
          className="flex min-h-[420px] flex-col overflow-hidden rounded-2xl border border-gray-100 bg-white"
        >
          <header className="flex items-center gap-3 border-b border-gray-100 px-5 py-4">
            <Avatar initials={activeConversation.initials} name={activeConversation.name} size="sm" />
            <div className="min-w-0">
              <p className="truncate text-sm font-bold text-gray-900">{activeConversation.name}</p>
              <p className="truncate text-xs text-gray-500">{activeConversation.farm}</p>
            </div>
          </header>

          <div className="flex-1 space-y-3 p-5">
            {messages.map((message, index) => (
              <div key={index} className={cn('flex', message.from === 'me' ? 'justify-end' : 'justify-start')}>
                <p
                  className={cn(
                    'max-w-[75%] rounded-2xl px-4 py-2.5 text-sm',
                    message.from === 'me'
                      ? 'rounded-br-md bg-brand text-white'
                      : 'rounded-bl-md bg-gray-100 text-gray-800',
                  )}
                >
                  {message.text}
                </p>
              </div>
            ))}
          </div>

          <form onSubmit={handleSend} className="flex items-center gap-2 border-t border-gray-100 p-3">
            <label htmlFor="chat-message" className="sr-only">
              Escribe un mensaje
            </label>
            <input
              id="chat-message"
              type="text"
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              placeholder="Escribe un mensaje..."
              className="flex-1 rounded-xl border border-gray-200 bg-gray-50 px-4 py-2.5 text-sm text-gray-900 placeholder:text-gray-400 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
            />
            <button
              type="submit"
              aria-label="Enviar mensaje"
              className="inline-flex items-center justify-center rounded-xl bg-brand p-2.5 text-white transition-colors hover:bg-brand-dark focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand"
            >
              <Icon name="send" size={18} />
            </button>
          </form>
        </section>
      </div>
    </div>
  )
}
