import { getToken } from '../lib/session'
import { chatUrl, request } from './http'

export const conversations = {
  // buyer_id is resolved from the authenticated principal server-side, so it is
  // deliberately not sent.
  create: (data) =>
    request('/conversations', { method: 'POST', body: data }),

  list: () =>
    request('/conversations'),

  getById: (id) => request(`/conversations/${id}`),

  messages: (id) => request(`/conversations/${id}/messages`),
}

/**
 * Opens the chat socket for a conversation.
 *
 * The token travels as a subprotocol because a browser WebSocket cannot set an
 * Authorization header. `milpa.chat.v1` has to be offered as well: the server
 * only negotiates that subprotocol, so a socket opened without it is rejected.
 *
 * @param {string} conversationId
 * @param {{onOpen?: () => void, onMessage?: (m: any) => void, onClose?: () => void, onError?: (e: any) => void}} handlers
 * @returns {WebSocket}
 */
export function openChat(conversationId, handlers = {}) {
  const url = chatUrl(conversationId)
  const token = getToken()

  if (!token) {
    throw new Error('Sesión expirada')
  }

  const socket = new WebSocket(url, ['milpa.chat.v1', `bearer.${token}`])

  socket.onopen = () => handlers.onOpen?.()
  socket.onmessage = (event) => {
    try {
      handlers.onMessage?.(JSON.parse(event.data))
    } catch {
      handlers.onMessage?.({ content: event.data })
    }
  }
  socket.onclose = () => handlers.onClose?.()
  socket.onerror = (event) => handlers.onError?.(event)

  return socket
}
