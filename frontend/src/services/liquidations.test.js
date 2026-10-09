import { beforeEach, describe, expect, it, vi } from 'vitest'
import { liquidations as liquidationsApi } from './liquidations'

function jsonResponse(status, body) {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
  }
}

describe('services/liquidations', () => {
  let fetchMock

  beforeEach(() => {
    fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
    localStorage.setItem('milpa_token', 'tok-123')
  })

  it('lista las de un proveedor con su supplier_id', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { data: [] }))

    await liquidationsApi.getBySupplier('u-1')

    expect(fetchMock.mock.calls[0][0]).toContain('/liquidations/?supplier_id=u-1')
    expect(fetchMock.mock.calls[0][1].method).toBeUndefined()
  })

  it('crea con POST y borra con DELETE', async () => {
    fetchMock.mockResolvedValue(jsonResponse(201, { data: { id: 'l2' } }))

    await liquidationsApi.create({ product_name: 'Maíz' })
    const [createUrl, createConfig] = fetchMock.mock.calls[0]
    expect(createUrl).toContain('/liquidations/')
    expect(createConfig.method).toBe('POST')
    expect(createConfig.body).toBe('{"product_name":"Maíz"}')

    fetchMock.mockResolvedValue(jsonResponse(200, {}))
    await liquidationsApi.remove('l2')
    expect(fetchMock.mock.calls[1][1].method).toBe('DELETE')
  })

  it('edita con PATCH enviando solo lo que cambia', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, {}))

    await liquidationsApi.update('l2', { quantity: 25 })

    const [url, config] = fetchMock.mock.calls[0]
    expect(url).toContain('/liquidations/l2')
    expect(config.method).toBe('PATCH')
    expect(config.body).toBe('{"quantity":25}')
  })
})