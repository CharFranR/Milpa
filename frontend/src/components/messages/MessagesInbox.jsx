import { useState } from 'react'
import ChatPanel from '../chat/ChatPanel'
import EmptyState from '../ui/EmptyState'
import ConversationList from './ConversationList'

export default function MessagesInbox() {
  const [selected, setSelected] = useState(null)

  return (
    <div className="grid gap-4 lg:grid-cols-[minmax(0,20rem)_minmax(0,1fr)]">
      <ConversationList onSelect={setSelected} selectedId={selected?.id} />
      {selected ? (
        <ChatPanel
          conversationId={selected.id}
          title={selected.peerName}
          hint={`Escribe a ${selected.peerName}. Tu mensaje llegará al instante.`}
          className="h-full"
        />
      ) : (
        <EmptyState
          icon="chat_bubble"
          title="Selecciona una conversación"
          description="Elige una conversación de la lista para leer el historial y responder."
          className="h-full"
        />
      )}
    </div>
  )
}
