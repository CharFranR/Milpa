import { act, render, screen, waitFor, fireEvent } from '@testing-library/react'
import { describe, expect, it, vi, beforeEach } from 'vitest'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from '../../context/AuthContext'
import Login from './Login'

vi.mock('../../services/auth', () => ({
  auth: {
    login: vi.fn(),
  },
}))

function renderLogin(routes = []) {
  return render(
    <MemoryRouter initialEntries={['/login']}>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/producer" element={<div data-testid="producer">Producer</div>} />
          <Route path="/admin" element={<div data-testid="admin">Admin</div>} />
          <Route path="/dashboard" element={<div data-testid="dashboard">Dashboard</div>} />
          {routes}
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  )
}

async function fillAndSubmit(email, password, roleLabel) {
  if (roleLabel) {
    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: roleLabel }))
    })
  }
  await act(async () => {
    fireEvent.change(screen.getByLabelText('Correo electrónico'), { target: { value: email } })
    fireEvent.change(screen.getByLabelText('Contraseña'), { target: { value: password } })
    fireEvent.click(screen.getByRole('button', { name: /Ingresar como /i }))
  })
}

describe('components/auth/Login', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('redirige a /producer para rol 1 (agricultor)', async () => {
    const { auth } = await import('../../services/auth')
    auth.login.mockResolvedValue({
      access_token: 'tok-1',
      user: { id: 'u1', first_name: 'Ana', role: 1 },
    })

    renderLogin()
    await fillAndSubmit('ana@test.com', 'Secret123!', 'Agricultor')

    await waitFor(() => expect(screen.getByTestId('producer')).toBeInTheDocument())
  })

  it('redirige a /admin para rol 5 (admin)', async () => {
    const { auth } = await import('../../services/auth')
    auth.login.mockResolvedValue({
      access_token: 'tok-2',
      user: { id: 'u2', first_name: 'Admin', role: 5 },
    })

    renderLogin()
    await fillAndSubmit('admin@test.com', 'Secret123!', 'Agricultor')

    await waitFor(() => expect(screen.getByTestId('admin')).toBeInTheDocument())
  })

  it('redirige a /dashboard para rol 3 (mayorista)', async () => {
    const { auth } = await import('../../services/auth')
    auth.login.mockResolvedValue({
      access_token: 'tok-3',
      user: { id: 'u3', first_name: 'Mayorista', role: 3 },
    })

    renderLogin()
    await fillAndSubmit('mayorista@test.com', 'Secret123!', 'Comprador')

    await waitFor(() => expect(screen.getByTestId('dashboard')).toBeInTheDocument())
  })

  it('redirige a /dashboard para rol 2 (minorista)', async () => {
    const { auth } = await import('../../services/auth')
    auth.login.mockResolvedValue({
      access_token: 'tok-4',
      user: { id: 'u4', first_name: 'Minorista', role: 2 },
    })

    renderLogin()
    await fillAndSubmit('minorista@test.com', 'Secret123!', 'Comprador')

    await waitFor(() => expect(screen.getByTestId('dashboard')).toBeInTheDocument())
  })

  it('muestra error para rol 0 (pending)', async () => {
    const { auth } = await import('../../services/auth')
    auth.login.mockResolvedValue({
      access_token: 'tok-5',
      user: { id: 'u5', first_name: 'Pendiente', role: 0 },
    })

    renderLogin()
    await fillAndSubmit('pending@test.com', 'Secret123!', 'Agricultor')

    await waitFor(() => expect(screen.getByText(/pendiente de aprobación/i)).toBeInTheDocument())
    expect(screen.queryByTestId('producer')).not.toBeInTheDocument()
    expect(screen.queryByTestId('dashboard')).not.toBeInTheDocument()
    expect(screen.queryByTestId('admin')).not.toBeInTheDocument()
  })
})