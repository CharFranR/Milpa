import { act, render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { AuthProvider, useAuth } from './AuthContext'
import { LOGOUT_EVENT, getToken, getUser, setToken, setUser } from '../lib/session'

let current

function Probe() {
  current = useAuth()
  return <span>{current.user?.first_name || 'anon'}</span>
}

function renderAuth() {
  return render(
    <AuthProvider>
      <Probe />
    </AuthProvider>,
  )
}

describe('context/AuthContext', () => {
  it('lee la sesión inicial y normaliza un rol numérico del backend', () => {
    setToken('tok-1')
    setUser({ id: 'u1', first_name: 'Ana', role: 2 })

    renderAuth()

    expect(current.role).toBe('buyer')
    expect(current.token).toBe('tok-1')
    expect(current.isAuthenticated).toBe(true)
    expect(screen.getByText('Ana')).toBeInTheDocument()
  })

  it('login guarda el token y el usuario en el storage', () => {
    renderAuth()
    expect(current.isAuthenticated).toBe(false)

    act(() => {
      current.login('tok-2', { id: 'u2', first_name: 'Ben', role: 'producer' })
    })

    expect(getToken()).toBe('tok-2')
    expect(getUser().first_name).toBe('Ben')
    expect(current.role).toBe('producer')
    expect(current.isAuthenticated).toBe(true)
  })

  it('updateUser refresca el usuario del contexto', () => {
    setToken('tok-3')
    setUser({ id: 'u3', first_name: 'Carla', role: 'buyer' })

    renderAuth()

    act(() => {
      current.updateUser({ id: 'u3', first_name: 'Carla Nueva', role: 'buyer' })
    })

    expect(current.user.first_name).toBe('Carla Nueva')
    expect(getUser().first_name).toBe('Carla Nueva')
    expect(screen.getByText('Carla Nueva')).toBeInTheDocument()
  })

  it('logout limpia storage y estado', () => {
    setToken('tok-4')
    setUser({ id: 'u4', first_name: 'Diego', role: 'buyer' })

    renderAuth()
    expect(screen.getByText('Diego')).toBeInTheDocument()

    act(() => {
      current.logout()
    })

    expect(getToken()).toBeNull()
    expect(getUser()).toBeNull()
    expect(current.user).toBeNull()
    expect(current.role).toBeNull()
    expect(current.isAuthenticated).toBe(false)
    expect(screen.getByText('anon')).toBeInTheDocument()
  })

  it('responde al evento de cierre de sesión que dispara http.js en un 401', () => {
    setToken('tok-5')
    setUser({ id: 'u5', first_name: 'Elena', role: 'buyer' })

    renderAuth()

    act(() => {
      window.dispatchEvent(new Event(LOGOUT_EVENT))
    })

    expect(current.user).toBeNull()
    expect(current.isAuthenticated).toBe(false)
  })

  it('descarta un rol que no pertenece al frontend', () => {
    setToken('tok-6')
    setUser({ id: 'u6', first_name: 'Fabi', role: 'superadmin' })

    renderAuth()

    expect(current.user).toBeNull()
    expect(current.role).toBeNull()
    expect(current.isAuthenticated).toBe(false)
    expect(screen.getByText('anon')).toBeInTheDocument()
  })
})
