import { describe, expect, it } from 'vitest'
import { hasExpired, isOpen, remainingLabel, statusLabel, toLocalInputValue } from './liquidations'

const NOW = new Date('2026-10-04T12:00:00Z')

describe('lib/liquidations', () => {
  it('detecta el cierre por expires_at', () => {
    expect(hasExpired({ expires_at: '2026-10-04T11:00:00Z' }, NOW)).toBe(true)
    expect(hasExpired({ expires_at: '2026-10-05T11:00:00Z' }, NOW)).toBe(false)
    expect(hasExpired({ expires_at: null }, NOW)).toBe(false)
    expect(hasExpired({}, NOW)).toBe(false)
  })

  it('una liquidación abierta con fecha vencida queda como expirada', () => {
    const liq = { status: 0, expires_at: '2026-10-04T11:00:00Z' }
    expect(isOpen(liq, NOW)).toBe(false)
    expect(statusLabel(liq, NOW)).toBe('Expirada')
    expect(statusLabel({ status: 0 }, NOW)).toBe('Abierta')
    expect(statusLabel({ status: 1 }, NOW)).toBe('Cerrada')
    expect(statusLabel({ status: 3 }, NOW)).toBe('Asignada')
  })

  it('cuenta el tiempo restante en unidades legibles', () => {
    expect(remainingLabel('', NOW)).toBe('')
    expect(remainingLabel('2026-10-04T11:00:00Z', NOW)).toBe('Expirada')
    expect(remainingLabel('2026-10-04T14:30:00Z', NOW)).toBe('2 h 30 min')
    expect(remainingLabel('2026-10-06T14:30:00Z', NOW)).toBe('2 d 2 h')
    expect(remainingLabel('2026-10-04T12:20:00Z', NOW)).toBe('20 min')
  })

  it('convierte a lo que espera el input datetime-local', () => {
    expect(toLocalInputValue('')).toBe('')
    expect(toLocalInputValue(null)).toBe('')
    expect(toLocalInputValue('no-es-fecha')).toBe('')
    expect(toLocalInputValue('2026-10-04T14:05:00Z')).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$/)
  })
})
