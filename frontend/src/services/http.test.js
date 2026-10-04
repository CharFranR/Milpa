import { beforeEach, describe, expect, it, vi } from 'vitest'
import { LOGOUT_EVENT } from '../lib/session'
import { request } from './http'

function jsonResponse(status, body) {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
  }
}

describe('services/http', () => {
  let fetchMock

  beforeEach(() => {
    fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
  })

  it('desenvuelve el envelope data', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { data: { id: 7 } }))
    await expect(request('/things')).resolves.toEqual({ id: 7 })
    expect(fetchMock.mock.calls[0][0]).toContain('/things')
  })

  it('devuelve el body tal cual cuando no trae envelope', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, [{ id: 1 }]))
    await expect(request('/things')).resolves.toEqual([{ id: 1 }])
  })

  it('serializa el body y adjunta el token como Authorization', async () => {
    localStorage.setItem('milpa_token', 'tok-123')
    fetchMock.mockResolvedValue(jsonResponse(201, { data: { id: 1 } }))

    await request('/things', { method: 'POST', body: { name: 'Milpa' } })

    const [, config] = fetchMock.mock.calls[0]
    expect(config.headers.Authorization).toBe('Bearer tok-123')
    expect(config.body).toBe('{"name":"Milpa"}')
  })

  it('no serializa FormData', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { data: null }))
    const form = new FormData()
    form.append('file', 'x')

    await request('/upload', { method: 'POST', body: form })

    expect(fetchMock.mock.calls[0][1].body).toBe(form)
  })

  it('401 cierra la sesión, dispara el evento y manda a login', async () => {
    localStorage.setItem('milpa_token', 'tok-vencido')
    localStorage.setItem('milpa_user', JSON.stringify({ id: 'u1', role: 'buyer' }))
    fetchMock.mockResolvedValue(jsonResponse(401, { error: 'unauthorized' }))

    const notified = vi.fn()
    window.addEventListener(LOGOUT_EVENT, notified)

    await expect(request('/things')).rejects.toMatchObject({
      message: 'Sesión expirada',
      status: 401,
    })

    window.removeEventListener(LOGOUT_EVENT, notified)
    expect(notified).toHaveBeenCalledTimes(1)
    expect(localStorage.getItem('milpa_token')).toBeNull()
    expect(localStorage.getItem('milpa_user')).toBeNull()
    expect(window.location.hash).toBe('#/login')
  })

  it('401 sobre /auth/login no cierra la sesión', async () => {
    localStorage.setItem('milpa_token', 'tok-123')
    fetchMock.mockResolvedValue(jsonResponse(401, { error: 'Credenciales inválidas' }))

    await expect(request('/auth/login', { method: 'POST' })).rejects.toMatchObject({
      message: 'Credenciales inválidas',
      status: 401,
    })

    expect(localStorage.getItem('milpa_token')).toBe('tok-123')
    expect(window.location.hash).toBe('')
  })

  it('403 responde con el mensaje de permisos', async () => {
    fetchMock.mockResolvedValue(jsonResponse(403, { error: 'forbidden' }))

    await expect(request('/things')).rejects.toMatchObject({
      message: 'No tienes permisos para realizar esta acción',
      status: 403,
    })
  })

  it('5xx responde con el mensaje de servicio no disponible', async () => {
    fetchMock.mockResolvedValue(jsonResponse(500, { error: 'redis: connection refused' }))

    await expect(request('/things')).rejects.toMatchObject({
      message: 'Servicio no disponible, intenta de nuevo',
      status: 500,
    })
  })

  it('red caída responde como servicio no disponible', async () => {
    fetchMock.mockRejectedValue(new TypeError('Failed to fetch'))

    await expect(request('/things')).rejects.toMatchObject({
      message: 'Servicio no disponible, intenta de nuevo',
      status: 0,
    })
  })

  it('conserva los errores de validación del backend', async () => {
    fetchMock.mockResolvedValue(jsonResponse(400, { error: 'confirm_password: cannot be blank' }))

    await expect(request('/auth/register', { method: 'POST' })).rejects.toMatchObject({
      message: 'confirm_password: cannot be blank',
      status: 400,
    })
  })
})
