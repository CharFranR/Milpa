import { beforeEach, describe, expect, it, vi } from 'vitest'
import { supplyOffers } from './supplyOffers'

function jsonResponse(status, body) {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
  }
}

describe('services/supplyOffers', () => {
  let fetchMock

  beforeEach(() => {
    fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
    localStorage.setItem('milpa_token', 'tok-123')
  })

  it('lista las ofertas propias', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { data: [{ id: 'o1' }] }))

    await expect(supplyOffers.listMine()).resolves.toEqual([{ id: 'o1' }])
    expect(fetchMock.mock.calls[0][0]).toContain('/supply-offers/')
  })

  it('crea una oferta con los campos que espera el backend', async () => {
    fetchMock.mockResolvedValue(jsonResponse(201, { data: { id: 'o2' } }))

    const created = await supplyOffers.create({
      supply_request_id: 'r1',
      total_amount: 5000,
      price_per_unit: 250,
      measurement: 1,
      delivery_available: true,
      comments: 'Listo para despachar',
    })

    expect(created).toEqual({ id: 'o2' })
    const [url, config] = fetchMock.mock.calls[0]
    expect(url).toContain('/supply-offers/')
    expect(config.method).toBe('POST')
    expect(JSON.parse(config.body)).toMatchObject({
      supply_request_id: 'r1',
      total_amount: 5000,
      price_per_unit: 250,
      measurement: 1,
      delivery_available: true,
    })
  })

  it('lista las ofertas de una solicitud concreta', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { data: [] }))

    await supplyOffers.listByRequest('r9')

    expect(fetchMock.mock.calls[0][0]).toContain('/supply-offers/requests/r9')
    expect(fetchMock.mock.calls[0][1].method).toBeUndefined()
  })

  it('trae una oferta por id', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { data: { id: 'o3' } }))

    await expect(supplyOffers.getById('o3')).resolves.toEqual({ id: 'o3' })
    expect(fetchMock.mock.calls[0][0]).toContain('/supply-offers/o3')
  })

  it('actualiza con PATCH', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, {}))

    await supplyOffers.update('o4', { total_amount: 7000, price_per_unit: 300 })

    const [url, config] = fetchMock.mock.calls[0]
    expect(url).toContain('/supply-offers/o4')
    expect(config.method).toBe('PATCH')
    expect(config.body).toBe('{"total_amount":7000,"price_per_unit":300}')
  })

  it('retira con POST sin body', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, {}))

    await supplyOffers.withdraw('o5')

    expect(fetchMock.mock.calls[0][0]).toContain('/supply-offers/o5/withdraw')
    expect(fetchMock.mock.calls[0][1].method).toBe('POST')
  })

  it('propaga los conflictos de negocio (409)', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse(409, { error: 'insufficient amount: offer total amount exceeds the remaining amount of the supply request' }),
    )

    await expect(
      supplyOffers.create({ supply_request_id: 'r1', total_amount: 999999, price_per_unit: 1 }),
    ).rejects.toMatchObject({ status: 409 })
  })

  it('impide ofertar en la solicitud propia con 403', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse(403, { error: 'forbidden: cannot create an offer on your own supply request' }),
    )

    await expect(
      supplyOffers.create({ supply_request_id: 'r1', total_amount: 10, price_per_unit: 1 }),
    ).rejects.toMatchObject({
      status: 403,
      message: 'No tienes permisos para realizar esta acción',
    })
  })
})
