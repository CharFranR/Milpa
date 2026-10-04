import { beforeEach, describe, expect, it, vi } from 'vitest'

function jsonResponse(status, body) {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
  }
}

describe('services/categories', () => {
  let fetchMock

  beforeEach(() => {
    fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
    vi.resetModules()
    localStorage.setItem('milpa_token', 'tok-123')
  })

  it('devuelve el arreglo de /categories', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { data: [{ id: 'c1', name: 'Frutales' }] }))
    const { categories } = await import('./categories')

    await expect(categories.getAll()).resolves.toEqual([{ id: 'c1', name: 'Frutales' }])
    expect(fetchMock.mock.calls[0][0]).toContain('/categories')
  })

  it('comparte una sola petición entre varios consumidores', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { data: [{ id: 'c1' }] }))
    const { categories } = await import('./categories')

    const [first, second] = await Promise.all([categories.getAll(), categories.getAll()])

    expect(fetchMock).toHaveBeenCalledTimes(1)
    expect(first).toEqual(second)
  })

  it('vuelve a pedir la lista si la petición anterior falló', async () => {
    fetchMock
      .mockRejectedValueOnce(new Error('sin red'))
      .mockResolvedValueOnce(jsonResponse(200, { data: [{ id: 'c2' }] }))
    const { categories } = await import('./categories')

    await expect(categories.getAll()).rejects.toMatchObject({
      message: 'Servicio no disponible, intenta de nuevo',
    })
    await expect(categories.getAll()).resolves.toEqual([{ id: 'c2' }])
    expect(fetchMock).toHaveBeenCalledTimes(2)
  })
})
