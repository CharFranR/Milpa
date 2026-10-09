import { describe, expect, it } from 'vitest'
import { iconForCategory } from './categoryIcons'

describe('lib/categoryIcons', () => {
  it('resuelve por main_category (texto sin acentos)', () => {
    expect(iconForCategory({ main_category: 'frutales' })).toBe('local_florist')
    expect(iconForCategory({ main_category: 'citrusos' })).toBe('nutrition')
    expect(iconForCategory({ main_category: 'otros' })).toBe('shopping_basket')
  })

  it('resuelve por nombre con acentos y espacios', () => {
    expect(iconForCategory({ name: 'Cítricos' })).toBe('nutrition')
    expect(iconForCategory({ name: 'Granos Básicos' })).toBe('grain')
    expect(iconForCategory({ name: 'Hortalizas' })).toBe('grass')
  })

  it('cae al icono explícito y si no a uno genérico', () => {
    expect(iconForCategory({ name: 'Sin mapear', icon: 'storefront' })).toBe('storefront')
    expect(iconForCategory({ name: 'Sin mapear' })).toBe('category')
    expect(iconForCategory(null)).toBe('category')
  })
})
