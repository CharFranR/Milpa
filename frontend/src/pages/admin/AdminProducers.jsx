import { useState, useEffect } from 'react'
import ProducerCard from '../../components/product/ProducerCard'
import Button from '../../components/ui/Button'
import Icon from '../../components/ui/Icon'
import { admin } from '../../services/api'

const PRODUCER_ROLE = 1

async function fetchAllProducers() {
  const items = []
  let page = 1
  for (;;) {
    const res = await admin.listUsers({ page, pageSize: 100 })
    const batch = Array.isArray(res.items) ? res.items : []
    items.push(...batch)
    if (batch.length === 0 || items.length >= res.total) break
    page += 1
  }
  return items.filter((u) => u.role === PRODUCER_ROLE)
}

function toCard(u) {
  const location = [u.municipality, u.department].filter(Boolean).join(', ')
  return {
    id: u.id,
    name: `${u.first_name || ''} ${u.last_name || ''}`.trim() || u.email,
    location: location || undefined,
    active: !u.suspended_at,
    since: u.created_at ? new Date(u.created_at).getFullYear() : undefined,
  }
}

export default function AdminProducers() {
  const [producers, setProducers] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  function load() {
    setLoading(true)
    setError('')
    fetchAllProducers()
      .then((items) => setProducers(items.map(toCard)))
      .catch((err) => setError(err.message || 'Error al cargar productores.'))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  if (loading) {
    return (
      <div className="space-y-6">
        <h1 className="text-3xl font-bold text-gray-900">Productores</h1>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {[1, 2, 3].map((i) => (
            <div key={i} className="animate-pulse rounded-2xl border border-gray-100 bg-white p-5">
              <div className="flex items-center gap-4">
                <div className="h-16 w-16 rounded-full bg-gray-200" />
                <div className="flex-1 space-y-2">
                  <div className="h-4 w-32 rounded bg-gray-200" />
                  <div className="h-3 w-44 rounded bg-gray-200" />
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-6">
        <h1 className="text-3xl font-bold text-gray-900">Productores</h1>
        <div className="rounded-xl border border-red-200 bg-red-50 p-6 text-center">
          <Icon name="error" size={40} className="mx-auto text-red-400" />
          <p className="mt-3 text-sm text-red-700">{error}</p>
          <Button variant="outline" size="sm" onClick={load} className="mt-4">
            Reintentar
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold text-gray-900">Productores</h1>

      {producers.length === 0 ? (
        <div className="rounded-xl border border-gray-100 bg-white p-12 text-center">
          <Icon name="agriculture" size={48} className="mx-auto text-gray-300" />
          <h2 className="mt-4 text-lg font-semibold text-gray-900">Sin productores</h2>
          <p className="mt-2 text-sm text-gray-500">
            Todavía no hay cuentas con rol de agricultor registradas.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 max-w-full">
          {producers.map((prod) => (
            <ProducerCard
              key={prod.id}
              producer={prod}
            />
          ))}
        </div>
      )}
    </div>
  )
}
