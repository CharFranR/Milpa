import { describe, expect, it } from 'vitest'
import {
  LOGOUT_EVENT,
  clearSession,
  getCompanyId,
  getSessionRole,
  getToken,
  getUser,
  hasRole,
  isAuthenticated,
  setCompanyId,
  setToken,
  setUser,
} from './session'

describe('lib/session', () => {
  it('guarda y lee el token', () => {
    expect(getToken()).toBeNull()
    setToken('abc')
    expect(getToken()).toBe('abc')
  })

  it('hace roundtrip del usuario', () => {
    setUser({ id: 'u1', first_name: 'Ana', role: 'buyer' })
    expect(getUser()).toEqual({ id: 'u1', first_name: 'Ana', role: 'buyer' })
  })

  it('devuelve null si el usuario guardado no es JSON válido', () => {
    localStorage.setItem('milpa_user', '{rotto')
    expect(getUser()).toBeNull()
  })

  it('guarda el id de empresa y lo borra con setCompanyId(null)', () => {
    setCompanyId('c1')
    expect(getCompanyId()).toBe('c1')
    setCompanyId(null)
    expect(getCompanyId()).toBeNull()
  })

  it('expone rol y autenticación a partir del storage', () => {
    expect(isAuthenticated()).toBe(false)
    expect(getSessionRole()).toBeNull()

    setToken('tok')
    setUser({ id: 'u1', role: 'producer' })

    expect(isAuthenticated()).toBe(true)
    expect(getSessionRole()).toBe('producer')
    expect(hasRole('producer')).toBe(true)
    expect(hasRole('buyer')).toBe(false)
  })

  it('clearSession borra todo y notifica el cierre', () => {
    setToken('tok')
    setUser({ id: 'u1', role: 'buyer' })
    setCompanyId('c1')

    let notified = 0
    const listener = () => { notified++ }
    window.addEventListener(LOGOUT_EVENT, listener)

    clearSession()
    window.removeEventListener(LOGOUT_EVENT, listener)

    expect(notified).toBe(1)
    expect(getToken()).toBeNull()
    expect(getUser()).toBeNull()
    expect(getCompanyId()).toBeNull()
  })
})
