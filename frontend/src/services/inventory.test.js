import { beforeEach, describe, expect, it, vi } from 'vitest'
import { inventory, recommendations } from './inventory'

function jsonResponse(status, body) {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
  }
}

describe('services/inventory', () => {
  let fetchMock

  beforeEach(() => {
    fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
    localStorage.setItem('milpa_token', 'tok-123')
  })

  it('lista el inventario del proveedor', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { data: [{ id: 'i1' }] }))

    await expect(inventory.list()).resolves.toEqual([{ id: 'i1' }])
    expect(fetchMock.mock.calls[0][0]).toContain('/inventory/')
  })

  it('hace upsert con PUT y los tres campos', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { data: { id: 'i2' } }))

    const saved = await inventory.upsert({
      product_name: 'Urea',
      quantity: 1500,
      measurement: 0,
    })

    expect(saved).toEqual({ id: 'i2' })
    const [url, config] = fetchMock.mock.calls[0]
    expect(url).toContain('/inventory/')
    expect(config.method).toBe('PUT')
    expect(JSON.parse(config.body)).toEqual({
      product_name: 'Urea',
      quantity: 1500,
      measurement: 0,
    })
  })

  it('elimina con DELETE', async () => {
    fetchMock.mockResolvedValue(jsonResponse(204, null))

    await expect(inventory.remove('i3')).resolves.toBeNull()

    expect(fetchMock.mock.calls[0][0]).toContain('/inventory/i3')
    expect(fetchMock.mock.calls[0][1].method).toBe('DELETE')
  })

  it('consulta la disponibilidad con supplier_id y product_name', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse(200, {
        data: { supplier_id: 'u1', product_name: 'Urea', available_quantity: 1200 },
      }),
    )

    await expect(
      recommendations.availability('u1', 'Urea'),
    ).resolves.toEqual({ supplier_id: 'u1', product_name: 'Urea', available_quantity: 1200 })

    const url = fetchMock.mock.calls[0][0]
    expect(url).toContain('supplier_id=u1')
    expect(url).toContain('product_name=Urea')
  })

  it('propaga la validación del backend', async () => {
    fetchMock.mockResolvedValue(jsonResponse(400, { error: 'invalid input: product_name is required' }))

    await expect(inventory.upsert({ product_name: '  ', quantity: 1, measurement: 0 })).rejects.toMatchObject({
      status: 400,
      message: 'invalid input: product_name is required',
    })
  })
})
