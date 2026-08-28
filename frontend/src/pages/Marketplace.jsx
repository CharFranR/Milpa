import { useState } from 'react'
import Navbar from '../components/layout/Navbar'
import Footer from '../components/layout/Footer'
import Icon from '../components/ui/Icon'
import ProductCardGrid from '../components/product/ProductCardGrid'
import ProductCardList from '../components/product/ProductCardList'
import FiltersSidebar, { PRICE_LIMIT } from '../components/marketplace/FiltersSidebar'
import Pagination from '../components/marketplace/Pagination'
import EmptyResults from '../components/marketplace/EmptyResults'
import { producerById, products } from '../mocks/catalog'
import { cn } from '../lib/cn'

const SORT_OPTIONS = [
  { value: 'relevant', label: 'Más relevantes' },
  { value: 'rating', label: 'Mejor valorados' },
  { value: 'priceAsc', label: 'Precio: menor a mayor' },
  { value: 'priceDesc', label: 'Precio: mayor a menor' },
]

const DEFAULT_FILTERS = {
  category: 'all',
  maxPrice: PRICE_LIMIT,
  minRating: 0,
}

const PAGE_SIZE = 6

export default function Marketplace() {
  const [query, setQuery] = useState('')
  const [sort, setSort] = useState('relevant')
  const [filters, setFilters] = useState(DEFAULT_FILTERS)
  const [view, setView] = useState('grid')
  const [page, setPage] = useState(1)
  const [showFilters, setShowFilters] = useState(false)

  const visibleProducts = filterAndSortProducts(query, filters, sort)
  const totalPages = Math.ceil(visibleProducts.length / PAGE_SIZE)
  const paginatedProducts = paginateProducts(visibleProducts, page)

  function updateFilters(patch) {
    setFilters((current) => ({ ...current, ...patch }))
    setPage(1)
  }

  function clearAll() {
    setQuery('')
    setFilters(DEFAULT_FILTERS)
    setPage(1)
  }

  return (
    <div className="flex min-h-screen flex-col bg-gray-50">
      <Navbar />

      <main className="mx-auto w-full max-w-7xl flex-1 px-4 py-8 sm:px-6 lg:px-8">
        <nav aria-label="Ruta de navegación" className="flex items-center gap-1 text-sm text-gray-500">
          <a href="#/" className="hover:text-brand">
            Inicio
          </a>
          <Icon name="chevron_right" size={16} className="text-gray-300" />
          <span className="font-semibold text-gray-900">Marketplace</span>
        </nav>

        <header className="mt-4">
          <h1 className="text-3xl font-bold tracking-tight text-gray-900 sm:text-4xl">
            Marketplace
          </h1>
          <p className="mt-1 text-sm text-gray-500">
            <span className="font-semibold text-brand">{visibleProducts.length}</span> productos
            disponibles de productores locales
          </p>
        </header>

        <div className="mt-6 flex gap-3 sm:flex-row sm:items-center">
          <div className="relative flex-1">
            <Icon
              name="search"
              size={18}
              className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"
            />
            <label htmlFor="marketplace-search" className="sr-only">
              Buscar productos o productores
            </label>
            <input
              id="marketplace-search"
              type="text"
              value={query}
              onChange={(e) => {
                setQuery(e.target.value)
                setPage(1)
              }}
              placeholder="Buscar productos o productores..."
              className="w-full rounded-xl border border-gray-200 bg-white py-2.5 pl-10 pr-10 text-sm text-gray-900 placeholder:text-gray-400 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
            />
            {query && (
              <button
                type="button"
                onClick={() => {
                  setQuery('')
                  setPage(1)
                }}
                aria-label="Limpiar búsqueda"
                className="absolute right-2 top-1/2 -translate-y-1/2 rounded-full p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
              >
                <Icon name="close" size={14} />
              </button>
            )}
          </div>

          <ViewToggle view={view} onChange={setView} />

          <button
            type="button"
            onClick={() => setShowFilters((v) => !v)}
            aria-expanded={showFilters}
            aria-label={showFilters ? 'Ocultar filtros' : 'Mostrar filtros'}
            className="inline-flex items-center justify-center rounded-xl border border-gray-200 bg-white p-2.5 text-gray-600 transition-colors hover:border-brand/40 hover:text-brand lg:hidden"
          >
            <Icon name="tune" size={20} />
          </button>

          <label htmlFor="marketplace-sort" className="sr-only">
            Ordenar por
          </label>
          <select
            id="marketplace-sort"
            value={sort}
            onChange={(e) => {
              setSort(e.target.value)
              setPage(1)
            }}
            className="rounded-xl border border-gray-200 bg-white px-3 py-2.5 text-sm font-medium text-gray-700 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
          >
            {SORT_OPTIONS.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </div>

        <div className="mt-6 flex flex-col gap-6 pb-4 lg:flex-row">
          <aside className="shrink-0 lg:w-60">
            <div className="lg:sticky lg:top-20">
              <div className={showFilters ? '' : 'hidden lg:block'}>
                <FiltersSidebar filters={filters} onChange={updateFilters} onClear={clearAll} />
              </div>
            </div>
          </aside>

          <section aria-label="Resultados" className="min-w-0 flex-1">
            {visibleProducts.length === 0 ? (
              <EmptyResults onClear={clearAll} />
            ) : view === 'grid' ? (
              <div className="grid gap-5 sm:grid-cols-2 xl:grid-cols-3">
                {paginatedProducts.map((product) => (
                  <ProductCardGrid key={product.id} productId={product.id} />
                ))}
              </div>
            ) : (
              <div className="space-y-4">
                {paginatedProducts.map((product) => (
                  <ProductCardList key={product.id} productId={product.id} />
                ))}
              </div>
            )}

            {totalPages > 1 && (
              <Pagination currentPage={page} totalPages={totalPages} onPageChange={setPage} />
            )}
          </section>
        </div>
      </main>

      <Footer />
    </div>
  )
}

function ViewToggle({ view, onChange }) {
  return (
    <div
      role="group"
      aria-label="Modo de vista"
      className="hidden items-center rounded-xl border border-gray-200 bg-white p-1 sm:flex"
    >
      <button
        type="button"
        onClick={() => onChange('grid')}
        aria-pressed={view === 'grid'}
        aria-label="Vista de cuadrícula"
        className={cn(
          'rounded-lg p-1.5 transition-colors',
          view === 'grid' ? 'bg-brand-soft text-brand' : 'text-gray-400 hover:text-gray-600',
        )}
      >
        <Icon name="grid_view" size={20} />
      </button>
      <button
        type="button"
        onClick={() => onChange('list')}
        aria-pressed={view === 'list'}
        aria-label="Vista de lista"
        className={cn(
          'rounded-lg p-1.5 transition-colors',
          view === 'list' ? 'bg-brand-soft text-brand' : 'text-gray-400 hover:text-gray-600',
        )}
      >
        <Icon name="view_list" size={20} />
      </button>
    </div>
  )
}

function filterAndSortProducts(query, filters, sort) {
  const normalizedQuery = query.trim().toLowerCase()

  const matching = products.filter((product) => {
    if (!matchesQuery(product, normalizedQuery)) return false
    if (!matchesCategory(product, filters.category)) return false
    if (product.price > filters.maxPrice) return false
    if (product.rating < filters.minRating) return false
    return true
  })

  return sortProducts(matching, sort)
}

function matchesQuery(product, normalizedQuery) {
  if (!normalizedQuery) return true
  const producerName = producerById(product.producerId)?.name ?? ''
  const searchableText = `${product.name} ${producerName}`.toLowerCase()
  return searchableText.includes(normalizedQuery)
}

function matchesCategory(product, category) {
  return category === 'all' || product.categoryId === category
}

function paginateProducts(items, page) {
  const start = (page - 1) * PAGE_SIZE
  return items.slice(start, start + PAGE_SIZE)
}

function sortProducts(items, sort) {
  switch (sort) {
    case 'rating':
      return [...items].sort((a, b) => b.rating - a.rating)
    case 'priceAsc':
      return [...items].sort((a, b) => a.price - b.price)
    case 'priceDesc':
      return [...items].sort((a, b) => b.price - a.price)
    default:
      return items
  }
}
