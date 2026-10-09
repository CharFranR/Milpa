import { useState } from 'react'
import Button from '../ui/Button'
import { friendlyReportError } from '../../lib/reportMessages'

const MIN_LENGTH = 10
const MAX_LENGTH = 500

export default function ReportForm({ targetLabel = 'este contenido', onSubmit, onCancel }) {
  const [reason, setReason] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  async function handleSubmit(event) {
    event.preventDefault()
    setError('')

    const trimmed = reason.trim()
    if (trimmed.length < MIN_LENGTH) {
      setError(`El motivo debe tener al menos ${MIN_LENGTH} caracteres.`)
      return
    }
    if (trimmed.length > MAX_LENGTH) {
      setError(`El motivo no puede superar los ${MAX_LENGTH} caracteres.`)
      return
    }

    setBusy(true)
    try {
      await onSubmit({ reason: trimmed })
      setReason('')
    } catch (err) {
      setError(friendlyReportError(err))
    } finally {
      setBusy(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-3">
      <label htmlFor="report_reason" className="block text-xs font-semibold text-gray-600">
        Motivo de la denuncia
      </label>
      <textarea
        id="report_reason"
        rows={3}
        value={reason}
        onChange={(event) => setReason(event.target.value)}
        placeholder={`Explica por qué quieres reportar ${targetLabel} (mínimo ${MIN_LENGTH} caracteres)`}
        disabled={busy}
        className="w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2.5 text-sm text-gray-900 placeholder:text-gray-400 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand disabled:opacity-60"
      />
      <div className="flex flex-wrap items-center gap-3">
        <Button type="submit" variant="danger" disabled={busy}>
          {busy ? 'Enviando…' : 'Enviar denuncia'}
        </Button>
        {onCancel && (
          <button
            type="button"
            onClick={onCancel}
            disabled={busy}
            className="text-sm font-semibold text-gray-500 hover:text-gray-700"
          >
            Cancelar
          </button>
        )}
        <span className="text-xs text-gray-400">{reason.trim().length}/{MAX_LENGTH}</span>
      </div>
      {error && (
        <p role="alert" className="text-sm text-red-600">
          {error}
        </p>
      )}
    </form>
  )
}
