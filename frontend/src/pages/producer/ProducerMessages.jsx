import Avatar from '../../components/Avatar'
import Badge from '../../components/ui/Badge'
import { producerConversations } from '../../mocks/producer'

export default function ProducerMessages() {
  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold text-gray-900">Mensajes</h1>

      <div className="space-y-3">
        {producerConversations.map((conv) => (
          <article
            key={conv.id}
            className="rounded-xl border border-gray-100 bg-white p-4 hover:border-brand/40 hover:shadow-sm transition-all"
          >
            <div className="flex items-center gap-4">
              <Avatar
                initials={conv.buyer.avatar}
                name={conv.buyer.name}
                size="lg"
                tone={conv.buyer.tone}
              />
              <div className="min-w-0 flex-1">
                <div className="flex items-center justify-between gap-2">
                  <p className="font-semibold text-gray-900 truncate">{conv.buyer.name}</p>
                  <span className="text-sm text-gray-500 whitespace-nowrap">{conv.time}</span>
                </div>
                <p className="mt-0.5 text-sm text-gray-500 truncate">
                  {conv.buyer.farm}
                </p>
                <p className="mt-1 text-sm text-gray-600 truncate">"{conv.lastMessage}"</p>
              </div>
              {conv.unread > 0 && (
                <Badge tone="red" className="shrink-0">
                  {conv.unread}
                </Badge>
              )}
            </div>
          </article>
        ))}
      </div>
    </div>
  )
}