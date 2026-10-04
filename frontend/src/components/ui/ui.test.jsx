import { act, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import ConfirmDialog from './ConfirmDialog'
import EmptyState from './EmptyState'
import ErrorState from './ErrorState'
import Spinner from './Spinner'
import Toast from './Toast'

describe('components/ui', () => {
  afterEach(() => {
    vi.useRealTimers()
  })

  it('Spinner anuncia el estado de carga', () => {
    render(<Spinner label="Cargando" />)
    expect(screen.getByRole('status')).toHaveTextContent('Cargando')
  })

  it('ErrorState muestra el error y permite reintentar', () => {
    const onRetry = vi.fn()
    render(<ErrorState message="Servicio no disponible, intenta de nuevo" onRetry={onRetry} />)

    expect(screen.getByRole('alert')).toHaveTextContent('Servicio no disponible, intenta de nuevo')
    fireEvent.click(screen.getByRole('button', { name: /reintentar/i }))
    expect(onRetry).toHaveBeenCalledTimes(1)
  })

  it('ErrorState oculta el botón de reintento sin handler', () => {
    render(<ErrorState message="Algo pasó" />)
    expect(screen.queryByRole('button')).not.toBeInTheDocument()
  })

  it('EmptyState renderiza título, descripción y acción', () => {
    render(
      <EmptyState
        title="Sin solicitudes"
        description="Todavía no has publicado nada"
        action={<button type="button">Crear solicitud</button>}
      />,
    )

    expect(screen.getByText('Sin solicitudes')).toBeInTheDocument()
    expect(screen.getByText('Todavía no has publicado nada')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Crear solicitud' }))
  })

  it('ConfirmDialog no se pinta cuando está cerrado', () => {
    render(<ConfirmDialog open={false} onConfirm={() => {}} onCancel={() => {}} />)
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('ConfirmDialog confirma, cancela y cierra con Escape', () => {
    const onConfirm = vi.fn()
    const onCancel = vi.fn()
    render(
      <ConfirmDialog
        open
        title="¿Cancelar solicitud?"
        message="Se eliminará para siempre"
        confirmLabel="Sí, cancelar"
        onConfirm={onConfirm}
        onCancel={onCancel}
      />,
    )

    const dialog = screen.getByRole('dialog')
    expect(dialog).toHaveAttribute('aria-modal', 'true')
    expect(screen.getByText('¿Cancelar solicitud?')).toBeInTheDocument()
    expect(document.activeElement).toHaveTextContent('Cancelar')

    fireEvent.click(screen.getByRole('button', { name: 'Sí, cancelar' }))
    expect(onConfirm).toHaveBeenCalledTimes(1)

    fireEvent.keyDown(window, { key: 'Escape' })
    expect(onCancel).toHaveBeenCalledTimes(1)

    fireEvent.click(screen.getByRole('button', { name: 'Cancelar' }))
    expect(onCancel).toHaveBeenCalledTimes(2)
  })

  it('Toast cierra solo después de su duración', () => {
    vi.useFakeTimers()
    const onClose = vi.fn()
    render(<Toast message="Perfil actualizado" onClose={onClose} />)

    expect(screen.getByRole('status')).toHaveTextContent('Perfil actualizado')
    act(() => vi.advanceTimersByTime(3100))
    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('Toast se puede cerrar a mano y no se pinta vacío', () => {
    const onClose = vi.fn()
    const { rerender } = render(<Toast message="Guardado" onClose={onClose} />)

    fireEvent.click(screen.getByRole('button', { name: 'Cerrar aviso' }))
    expect(onClose).toHaveBeenCalledTimes(1)

    rerender(<Toast message="" onClose={onClose} />)
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })
})
