import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { AuthProvider } from '../../context/AuthContext'
import RequireRole from './RequireRole'

function renderAt(path) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<div>pagina de login</div>} />
          <Route path="/admin" element={<div>panel admin</div>} />
          <Route
            path="/dashboard"
            element={
              <RequireRole role="minorista">
                <div>zona de comprador</div>
              </RequireRole>
            }
          />
          <Route
            path="/producer"
            element={
              <RequireRole role="producer">
                <div>zona de productor</div>
              </RequireRole>
            }
          />
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  )
}

describe('components/auth/RequireRole', () => {
  it('deja pasar cuando el rol coincide', () => {
    localStorage.setItem('milpa_token', 'tok')
    localStorage.setItem('milpa_user', JSON.stringify({ id: 'u1', role: 'minorista' }))

    renderAt('/dashboard')

    expect(screen.getByText('zona de comprador')).toBeInTheDocument()
    expect(screen.queryByText('pagina de login')).not.toBeInTheDocument()
  })

  it('redirige a la zona del rol cuando el rol no coincide (usuario autenticado)', () => {
    localStorage.setItem('milpa_token', 'tok')
    localStorage.setItem('milpa_user', JSON.stringify({ id: 'u1', role: 'minorista' }))

    renderAt('/producer')

    expect(screen.getByText('zona de comprador')).toBeInTheDocument()
    expect(screen.queryByText('pagina de login')).not.toBeInTheDocument()
    expect(screen.queryByText('zona de productor')).not.toBeInTheDocument()
  })

  it('redirige a login sin sesión', () => {
    renderAt('/dashboard')

    expect(screen.getByText('pagina de login')).toBeInTheDocument()
    expect(screen.queryByText('zona de comprador')).not.toBeInTheDocument()
  })
})