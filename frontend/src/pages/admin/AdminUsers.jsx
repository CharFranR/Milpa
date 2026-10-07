import { useState, useEffect, useCallback } from 'react'
import { admin } from '../../services/api'
import { normalizeRole, ROLE_LABELS } from '../../lib/roles'
import Icon from '../../components/ui/Icon'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'

const ROLE_FILTER_OPTIONS = [
  { value: 'all', label: 'Todos los roles' },
  { value: 'producer', label: 'Agricultor' },
  { value: 'minorista', label: 'Comprador Minorista' },
  { value: 'mayorista', label: 'Comprador Mayorista' },
  { value: 'admin', label: 'Administrador' },
  { value: 'auditor', label: 'Auditor' },
  { value: 'pending', label: 'Pendiente' },
]

export default function AdminUsers() {
  const [users, setUsers] = useState([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize] = useState(20)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [search, setSearch] = useState('')
  const [roleFilter, setRoleFilter] = useState('all')

  const fetchUsers = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await admin.listUsers({ page, pageSize })
      setUsers(data.items || [])
      setTotal(data.total || 0)
    } catch (e) {
      setError('No se pudieron cargar los usuarios')
      console.error(e)
    } finally {
      setLoading(false)
    }
  }, [page, pageSize])

  useEffect(() => {
    fetchUsers()
  }, [fetchUsers])

  const handleSearch = (e) => {
    setSearch(e.target.value)
    setPage(1)
  }

  const handleRoleFilter = (e) => {
    setRoleFilter(e.target.value)
    setPage(1)
  }

  const filteredUsers = users.filter((u) => {
    const role = normalizeRole(u.role)
    const matchesSearch = u.first_name?.toLowerCase().includes(search.toLowerCase()) ||
      u.last_name?.toLowerCase().includes(search.toLowerCase()) ||
      u.email?.toLowerCase().includes(search.toLowerCase())
    const matchesRole = roleFilter === 'all' || role === roleFilter
    return matchesSearch && matchesRole
  })

  const handleSuspend = async (id, action) => {
    try {
      await admin.suspendUser(id, action)
      fetchUsers()
    } catch (e) {
      alert('Error al suspender/reactivar usuario')
      console.error(e)
    }
  }

  const getStatusBadge = (user) => {
    if (user.suspended_at) return <Badge tone="amber">Suspendido</Badge>
    if (normalizeRole(user.role) === 'pending') return <Badge tone="amber">Pendiente</Badge>
    return <Badge tone="brand">Activo</Badge>
  }

  const getRoleBadge = (role) => {
    const normalized = normalizeRole(role)
    const _label = ROLE_LABELS[normalized] || 'Desconocido'
    const tones = { producer: 'brand', minorista: 'blue', mayorista: 'blue', admin: 'red', auditor: 'amber', pending: 'amber' }
    return <Badge tone={tones[normalized] || 'gray'}>{_label}</Badge>
  }

  if (loading) {
    return (
      <div className="space-y-4">
        <div className="h-8 bg-gray-100 rounded w-1/4 animate-pulse" />
        <div className="h-64 bg-gray-100 rounded animate-pulse" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="text-center py-12">
        <Icon name="error" size={48} className="mx-auto text-red-500 mb-4" />
        <p className="text-gray-600">{error}</p>
        <Button onClick={fetchUsers} className="mt-4">Reintentar</Button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900">Usuarios</h1>
      </div>

      <div className="flex flex-col gap-4 sm:flex-row">
        <div className="relative flex-1">
          <Icon name="search" size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
          <input
            type="text"
            value={search}
            onChange={handleSearch}
            placeholder="Buscar usuario..."
            aria-label="Buscar usuario"
            className="w-full rounded-xl border border-gray-200 bg-white py-2.5 pl-10 pr-4 text-sm text-gray-900 placeholder:text-gray-400 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
          />
        </div>
        <select
          value={roleFilter}
          onChange={handleRoleFilter}
          aria-label="Filtrar por rol"
          className="rounded-xl border border-gray-200 bg-white px-4 py-2.5 text-sm font-medium text-gray-700 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
        >
          {ROLE_FILTER_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
      </div>

      <div className="rounded-xl border border-gray-100 bg-white overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm" role="table">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Usuario</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Correo</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Tipo</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Región</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Miembro desde</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Estado</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-500">Acciones</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {filteredUsers.length === 0 ? (
                <tr>
                  <td colSpan={7} className="px-4 py-12 text-center text-gray-500">
                    {search || roleFilter !== 'all' ? 'No hay usuarios que coincidan' : 'No hay usuarios registrados'}
                  </td>
                </tr>
              ) : (
                filteredUsers.map((user) => (
                  <tr key={user.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <span className="inline-flex h-7 w-7 items-center justify-center rounded-full bg-brand-soft text-brand text-xs font-bold">
                          {(user.first_name?.[0] || '') + (user.last_name?.[0] || '')}
                        </span>
                        <span className="font-medium text-gray-900">
                          {user.first_name} {user.last_name}
                        </span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-gray-500">{user.email}</td>
                    <td className="px-4 py-3">{getRoleBadge(user.role)}</td>
                    <td className="px-4 py-3 text-gray-500">
                      {user.municipality} / {user.department}
                    </td>
                    <td className="px-4 py-3 text-gray-500">
                      {new Date(user.created_at).toLocaleDateString('es-ES', { year: 'numeric', month: 'short' })}
                    </td>
                    <td className="px-4 py-3">{getStatusBadge(user)}</td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <Button variant="outline" size="sm" onClick={() => {}}>Ver</Button>
                        {user.suspended_at ? (
                          <Button variant="primary" size="sm" onClick={() => handleSuspend(user.id, 'reactivate')}>
                            Reactivar
                          </Button>
                        ) : (
                          <Button variant="outline" size="sm" onClick={() => handleSuspend(user.id, 'suspend')}>
                            Suspender
                          </Button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
        {total > pageSize && (
          <div className="px-4 py-3 border-t border-gray-100 flex items-center justify-between">
            <span className="text-sm text-gray-500">
              Mostrando {(page - 1) * pageSize + 1}–{Math.min(page * pageSize, total)} de {total}
            </span>
            <div className="flex gap-2">
              <Button variant="outline" size="sm" onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page === 1}>Anterior</Button>
              <Button variant="outline" size="sm" onClick={() => setPage(p => p + 1)} disabled={page * pageSize >= total}>Siguiente</Button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}