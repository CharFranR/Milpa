import MessagesInbox from '../../components/messages/MessagesInbox'

export default function ProducerMessages() {
  return (
    <div className="space-y-6">
      <header>
        <h1 className="text-3xl font-bold text-gray-900">Mensajes</h1>
        <p className="mt-1 text-sm text-gray-500">
          Conversaciones con compradores: consultas de producto y chats de match.
        </p>
      </header>

      <MessagesInbox />
    </div>
  )
}
