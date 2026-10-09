import { useCallback, useEffect, useState } from 'react'
import { inventory, recommendations } from '../services/inventory'

export function useInventory() {
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const reload = useCallback(() => {
    setLoading(true)
    setError('')
    inventory
      .list()
      .then((data) => setItems(Array.isArray(data) ? data : []))
      .catch((err) => setError(err.message || 'Error al cargar el inventario.'))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    reload()
  }, [reload])

  return { items, loading, error, reload }
}

export function useAvailability(supplierId, items) {
  const key = items.map((item) => `${item.product_name} ${item.updated_at}`).join('|')
  const [map, setMap] = useState({})

  useEffect(() => {
    const names = items.map((item) => item.product_name)
    if (!supplierId || names.length === 0) {
      setMap({})
      return undefined
    }
    let alive = true
    Promise.all(
      names.map((productName) =>
        recommendations
          .availability(supplierId, productName)
          .then((data) => [productName, data.available_quantity])
          .catch(() => [productName, null]),
      ),
    ).then((pairs) => {
      if (alive) setMap(Object.fromEntries(pairs))
    })
    return () => {
      alive = false
    }
  }, [supplierId, key, items])

  return map
}
