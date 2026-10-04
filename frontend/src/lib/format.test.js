import { describe, expect, it } from 'vitest'
import { formatPrice, money } from './format'

describe('lib/format', () => {
  it('formatea precios en córdobas con separador de miles', () => {
    expect(formatPrice(1500)).toBe('C$1,500')
    expect(formatPrice(0)).toBe('C$0')
  })

  it('money devuelve el valor como texto', () => {
    expect(money(1500)).toBe('1500')
  })
})
