import { useEffect, useState } from 'react'
import SectionHeading from '../../components/SectionHeading'
import CategoryPill from '../../components/CategoryPill'
import { categories } from '../../services/categories'
import { iconForCategory } from '../../lib/categoryIcons'
import { Link } from 'react-router-dom'

export default function Categories() {
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancelled = false

    categories
      .getAll()
      .then((list) => {
        if (!cancelled) setItems(Array.isArray(list) ? list : [])
      })
      .catch(() => {
        if (!cancelled) setItems([])
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [])

  const visible = items.filter((category) => category.is_active !== false)

  if (!loading && visible.length === 0) return null

  return (
    <section className="bg-gray-50">
      <div className="mx-auto max-w-7xl px-4 py-16 sm:px-6 lg:px-8">
        <SectionHeading
          eyebrow="Explora"
          title="Categorías principales"
          action={
            <Link
              to="/marketplace"
              className="group inline-flex items-center gap-1 text-sm font-semibold text-brand hover:text-brand-dark"
            >
              Ver todas
              <span aria-hidden="true" className="transition-transform group-hover:translate-x-0.5">
                →
              </span>
            </Link>
          }
        />
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
          {loading
            ? Array.from({ length: 4 }).map((_, index) => (
                <div key={index} className="h-11 animate-pulse rounded-2xl bg-gray-200" />
              ))
            : visible.map((category) => (
                <CategoryPill
                  key={category.id}
                  icon={iconForCategory(category)}
                  name={category.name}
                />
              ))}
        </div>
      </div>
    </section>
  )
}
