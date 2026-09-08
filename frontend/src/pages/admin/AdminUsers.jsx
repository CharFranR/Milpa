import { useState } from 'react'
import { cn } from '../../lib/cn'
import Icon from '../../components/ui/Icon'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'
import { adminUsers } from '../../mocks/admin'

export default function AdminUsers() {
  const [search, setSearch] = useState('')
  const [roleFilter, setRoleFilter] = useState('all')

  const filteredUsers = adminUsers.filter((u) => {
    const matchesSearch = u.name.toLowerCase().includes(search.toLowerCase()) ||
      u.email.toLowerCase().includes(search.toLowerCase())
    const matchesRole = roleFilter === 'all' || u.type === roleFilter
    return matchesSearch && matchesRole
  })

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
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Buscar usuario..."
            className="w-full rounded-xl border border-gray-200 bg-white py-2.5 pl-10 pr-4 text-sm text-gray-900 placeholder:text-gray-400 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
          />
        </div>
        <select
          value={roleFilter}
          onChange={(e) => setRoleFilter(e.target.value)}
          className="rounded-xl border border-gray-200 bg-white px-4 py-2.5 text-sm font-medium text-gray-700 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
        >
          <option value="all">Todos los roles</option>
          <option value="Comprador">Comprador</option>
          <option value="Productor">Productor</option>
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
              {filteredUsers.map((user, i) => (
                <tr key={i} className="hover:bg-gray-50">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <span className="inline-flex h-7 w-7 items-center justify-center rounded-full bg-brand-soft text-brand text-xs font-bold">
                        {user.name.split(' ').map((n) => n[0]).join('').slice(0, 2).toUpperCase()}
                      </span>
                      <span className="font-medium text-gray-900">{user.name}</span>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-gray-500">{user.email}</td>
                  <td className="px-4 py-3">
                    <Badge tone={user.type === 'Productor' ? 'brand' : 'blue'}>
                      {user.type}
                    </Badge>
                  </td>
                  <td className="px-4 py-3 text-gray-500">{user.region}</td>
                  <td className="px-4 py-3 text-gray-500">{user.since}</td>
                  <td className="px-4 py-3">
                    <Badge tone={user.status === 'Activo' ? 'brand' : 'amber'}>
                      {user.status}
                    </Badge>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <Button variant="outline" size="sm">Ver</Button>
                      <Button
                        variant={user.status === 'Pendiente' ? 'primary' : 'outline'}
                        size="sm"
                        className={user.status === 'Pendiente' ? 'text-red-600 hover:bg-red-50' : 'text-gray-600 hover:bg-red-50'}
                      >
                        {user.status === 'Pendiente' ? 'Aprobar' : 'Bloquear'}
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}