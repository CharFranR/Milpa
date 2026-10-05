import { beforeEach, describe, expect, it, vi } from 'vitest'
import { transactions } from './transactions'

function jsonResponse(status, body) {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
  }
}

describe('services/transactions', () => {
  let fetchMock

  beforeEach(() => {
    fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
    localStorage.setItem('milpa_token', 'tok-123')
  })

  it('trae la transacción de un match', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { data: { id: 't1', status: 0 } }))

    await expect(transactions.getByMatch('m1')).resolves.toEqual({ id: 't1', status: 0 })
    expect(fetchMock.mock.calls[0][0]).toContain('/transactions/matches/m1')
    expect(fetchMock.mock.calls[0][1].method).toBeUndefined()
  })

  it('lista las transacciones de una solicitud', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { data: [{ id: 't2' }] }))

    await expect(transactions.listByRequest('r1')).resolves.toEqual([{ id: 't2' }])
    expect(fetchMock.mock.calls[0][0]).toContain('/transactions/requests/r1')
  })

  it('confirma inicio con POST sin body', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { data: null }))

    await transactions.confirmStart('t3')

    expect(fetchMock.mock.calls[0][0]).toContain('/transactions/t3/confirm-start')
    expect(fetchMock.mock.calls[0][1].method).toBe('POST')
  })

  it('confirma entrega con POST sin body', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { data: null }))

    await transactions.confirmDelivery('t4')

    expect(fetchMock.mock.calls[0][0]).toContain('/transactions/t4/confirm-delivery')
    expect(fetchMock.mock.calls[0][1].method).toBe('POST')
  })

  it('cancela enviando el motivo', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { data: null }))

    await transactions.cancel('t5', 'El proveedor no despachó')

    const [url, config] = fetchMock.mock.calls[0]
    expect(url).toContain('/transactions/t5/cancel')
    expect(config.method).toBe('POST')
    expect(JSON.parse(config.body)).toEqual({ reason: 'El proveedor no despachó' })
  })

  it('propaga los conflictos de transición (409)', async () => {
    fetchMock.mockResolvedValue(jsonResponse(409, { error: 'invalid transaction transition' }))

    await expect(transactions.confirmDelivery('t6')).rejects.toMatchObject({
      status: 409,
      message: 'invalid transaction transition',
    })
  })
})
