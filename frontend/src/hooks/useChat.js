import { useCallback, useEffect, useRef, useState } from 'react'
import { conversations, openChat } from '../services/conversations'

const MAX_DELAY = 12000

export function useChat(conversationId) {
  const [messages, setMessages] = useState([])
  const [status, setStatus] = useState('idle')
  const [error, setError] = useState('')
  const socketRef = useRef(null)
  const timerRef = useRef(null)
  const attemptRef = useRef(0)
  const closedRef = useRef(true)
  const connectRef = useRef(null)

  function scheduleReconnect() {
    if (closedRef.current || timerRef.current) return
    const delay = Math.min(800 * 2 ** attemptRef.current, MAX_DELAY)
    attemptRef.current += 1
    timerRef.current = setTimeout(() => {
      timerRef.current = null
      connectRef.current?.()
    }, delay)
  }

  function connect() {
    if (!conversationId || closedRef.current) return

    setStatus((prev) => (prev === 'connected' ? 'connected' : attemptRef.current > 0 ? 'reconnecting' : 'connecting'))

    let socket
    try {
      socket = openChat(conversationId, {
        onOpen: () => {
          attemptRef.current = 0
          setStatus('connected')
          setError('')
        },
        onMessage: (message) => {
          setMessages((prev) =>
            message?.id && prev.some((item) => item.id === message.id) ? prev : [...prev, message],
          )
        },
        onClose: () => {
          if (closedRef.current) return
          setStatus('reconnecting')
          scheduleReconnect()
        },
        onError: () => {
          setError('No se pudo conectar el chat.')
        },
      })
    } catch (err) {
      setStatus('error')
      setError(err?.message || 'No se pudo conectar el chat.')
      return
    }

    socketRef.current = socket
  }

  const reloadHistory = useCallback(() => {
    if (!conversationId) return
    setError('')
    conversations
      .messages(conversationId)
      .then((data) => {
        const history = Array.isArray(data) ? data : []
        setMessages((prev) => {
          const live = prev.filter((item) => item.id && !history.some((entry) => entry.id === item.id))
          return [...history, ...live]
        })
      })
      .catch((err) => setError(err.message || 'No se pudo cargar el historial.'))
  }, [conversationId])

  useEffect(() => {
    connectRef.current = connect
  })

  useEffect(() => {
    if (!conversationId) {
      setMessages([])
      setStatus('idle')
      return undefined
    }

    closedRef.current = false
    attemptRef.current = 0
    setStatus('loading')
    setMessages([])
    reloadHistory()
    connect()

    return () => {
      closedRef.current = true
      if (timerRef.current) {
        clearTimeout(timerRef.current)
        timerRef.current = null
      }
      socketRef.current?.close()
      socketRef.current = null
      setStatus('idle')
    }
  }, [conversationId, reloadHistory])

  const send = useCallback((content) => {
    const socket = socketRef.current
    const text = String(content || '').trim()
    if (!text) return false
    if (!socket || socket.readyState !== WebSocket.OPEN) return false
    socket.send(text)
    return true
  }, [])

  return { messages, status, error, send, reloadHistory }
}
