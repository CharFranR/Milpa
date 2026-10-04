import { beforeEach, describe, expect, it, vi } from 'vitest'
import { matches } from './matches'

function jsonResponse(status, body) {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
  }
}

describe('services/matches', () => {
  let fetchMock

  beforeEach(() => {
    fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
    localStorage.setItem('milpa_token', 'tok-123')
  })

  it('hace like a una oferta y devuelve match + transacción', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse(201, { data: { match: { id: 'm1' }, transaction: { id: 't1', status: 0 } } }),
    )

    await expect(matches.like('o1')).resolves.toEqual({
      match: { id: 'm1' },
      transaction: { id: 't1', status: 0 },
    })

    const [url, config] = fetchMock.mock.calls[0]
    expect(url).toContain('/matches/like/o1')
    expect(config.method).toBe('POST')
  })

  it('descarta con pass', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { data: null }))

    await expect(matches.pass('o2')).resolves.toBeNull()
    expect(fetchMock.mock.calls[0][0]).toContain('/matches/pass/o2')
    expect(fetchMock.mock.calls[0][1].method).toBe('POST')
  })

  it('lista los matches de una solicitud', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { data: [{ id: 'm2' }] }))

    await expect(matches.listByRequest('r1')).resolves.toEqual([{ id: 'm2' }])
    expect(fetchMock.mock.calls[0][0]).toContain('/matches/requests/r1')
    expect(fetchMock.mock.calls[0][1].method).toBeUndefined()
  })

  it('trae la lista priorizada con score y contribuciones', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse(200, {
        data: [
          {
            offer: { id: 'o3', supplier_id: 'u1' },
            score: 4.5,
            available_quantity: 800,
            contributions: [{ factor: 'price', weight: 0.6, score: 5, weighted_score: 3 }],
          },
        ],
      }),
    )

    const ranked = await matches.prioritized('r1')
    expect(ranked[0].score).toBe(4.5)
    expect(ranked[0].contributions[0].factor).toBe('price')
    expect(fetchMock.mock.calls[0][0]).toContain('/matches/requests/r1/prioritized')
  })

  it('trae un match por id', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { data: { id: 'm3', status: 0 } }))

    await expect(matches.getById('m3')).resolves.toEqual({ id: 'm3', status: 0 })
    expect(fetchMock.mock.calls[0][0]).toContain('/matches/m3')
  })

  it('propaga los conflictos de like (409)', async () => {
    fetchMock.mockResolvedValue(jsonResponse(409, { error: 'invalid match status' }))

    await expect(matches.like('o4')).rejects.toMatchObject({
      status: 409,
      message: 'invalid match status',
    })
  })

  it('impide operar sobre una solicitud ajena (403)', async () => {
    fetchMock.mockResolvedValue(jsonResponse(403, { error: 'forbidden' }))

    await expect(matches.prioritized('r2')).rejects.toMatchObject({
      status: 403,
      message: 'No tienes permisos para realizar esta acción',
    })
  })
})
