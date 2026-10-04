import { beforeEach, describe, expect, it, vi } from 'vitest'
import { supplyRequests } from './supplyRequests'

function jsonResponse(status, body) {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
  }
}

describe('services/supplyRequests', () => {
  let fetchMock

  beforeEach(() => {
    fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
    localStorage.setItem('milpa_token', 'tok-123')
  })

  it('lista las solicitudes propias', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { data: [{ id: 's1' }] }))

    await expect(supplyRequests.list()).resolves.toEqual([{ id: 's1' }])
    expect(fetchMock.mock.calls[0][0]).toContain('/supply-requests/')
    expect(fetchMock.mock.calls[0][1].method).toBeUndefined()
  })

  it('crea una solicitud con POST y desenvuelve el envelope', async () => {
    fetchMock.mockResolvedValue(jsonResponse(201, { data: { id: 's2' } }))

    const created = await supplyRequests.create({ product_name: 'Urea', total_amount: 5000 })

    expect(created).toEqual({ id: 's2' })
    const [url, config] = fetchMock.mock.calls[0]
    expect(url).toContain('/supply-requests/')
    expect(config.method).toBe('POST')
    expect(config.body).toBe('{"product_name":"Urea","total_amount":5000}')
    expect(config.headers.Authorization).toBe('Bearer tok-123')
  })

  it('trae una solicitud por id', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { data: { id: 's3' } }))

    await expect(supplyRequests.getById('s3')).resolves.toEqual({ id: 's3' })
    expect(fetchMock.mock.calls[0][0]).toContain('/supply-requests/s3')
  })

  it('lista las solicitudes disponibles para ofrecer', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { data: [] }))

    await supplyRequests.available()

    expect(fetchMock.mock.calls[0][0]).toContain('/supply-requests/available')
  })

  it('actualiza con PATCH enviando el payload completo', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, {}))

    await supplyRequests.update('s4', { product_name: 'Arroz', total_amount: 1000 })

    const [url, config] = fetchMock.mock.calls[0]
    expect(url).toContain('/supply-requests/s4')
    expect(config.method).toBe('PATCH')
    expect(config.body).toBe('{"product_name":"Arroz","total_amount":1000}')
  })

  it('usa los sub-recursos de montos y plazos', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, {}))

    await supplyRequests.updateAmounts('s5', { total_amount: 200 })
    await supplyRequests.updateDeadlines('s5', { delivery_deadline: '2026-11-01T00:00:00Z' })

    expect(fetchMock.mock.calls[0][0]).toContain('/supply-requests/s5/amounts')
    expect(fetchMock.mock.calls[0][1].method).toBe('PATCH')
    expect(fetchMock.mock.calls[1][0]).toContain('/supply-requests/s5/deadlines')
    expect(fetchMock.mock.calls[1][1].method).toBe('PATCH')
  })

  it('cancela y expira con POST sin body', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, {}))

    await supplyRequests.cancel('s6')
    await supplyRequests.expire('s7')

    expect(fetchMock.mock.calls[0][0]).toContain('/supply-requests/s6/cancel')
    expect(fetchMock.mock.calls[0][1].method).toBe('POST')
    expect(fetchMock.mock.calls[1][0]).toContain('/supply-requests/s7/expire')
    expect(fetchMock.mock.calls[1][1].method).toBe('POST')
  })

  it('propaga los conflictos 409 con su mensaje', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse(409, { error: 'supply request already has an active match' }),
    )

    await expect(supplyRequests.cancel('s8')).rejects.toMatchObject({
      message: 'supply request already has an active match',
      status: 409,
    })
  })

  it('conserva los errores de validación del backend', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse(400, { error: 'invalid input: product name is required' }),
    )

    await expect(
      supplyRequests.create({ product_name: '', total_amount: 10 }),
    ).rejects.toMatchObject({
      message: 'invalid input: product name is required',
      status: 400,
    })
  })
})
