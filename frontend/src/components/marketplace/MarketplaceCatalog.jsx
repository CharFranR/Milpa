import { useState } from 'react'
import Icon from '../ui/Icon'
import ProductCardGrid from '../../components/product/ProductCardGrid'
import ProductCardList from '../../components/product/ProductCardList'
import FiltersSidebar, { PRICE_LIMIT } from './FiltersSidebar'
import Pagination from './Pagination'
import EmptyResults from './EmptyResults'
import { useFeaturedOfferings } from '../../hooks/useFeaturedOfferings'
import { cn } from '../../lib/cn'

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

export default function MarketplaceCatalog() {
  const { offeringsList, loading, error } = useFeaturedOfferings()
  const [query, setQuery] = useState('')
  const [sort, setSort] = useState('relevant')
  const [filters, setFilters] = useState(DEFAULT_FILTERS)
  const [view, setView] = useState('grid')
  const [page, setPage] = useState(1)
  const [showFilters, setShowFilters] = useState(false)

  const visibleProducts = filterAndSortProducts(offeringsList, query, filters, sort)
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

  if (loading) {
    return (
      <>
        <header className="mt-4">
          <h1 className="text-3xl font-bold tracking-tight text-gray-900 sm:text-4xl">
            Marketplace
          </h1>
          <p className="mt-1 text-sm text-gray-500">Cargando productos...</p>
        </header>
        <div className="mt-6 grid gap-5 sm:grid-cols-2 xl:grid-cols-3">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <div key={i} className="animate-pulse rounded-2xl border border-gray-100 bg-white p-4 space-y-3">
              <div className="aspect-[4/3] rounded-xl bg-gray-200" />
              <div className="h-4 w-3/4 rounded bg-gray-200" />
              <div className="h-3 w-1/2 rounded bg-gray-200" />
              <div className="h-5 w-1/3 rounded bg-gray-200" />
            </div>
          ))}
        </div>
      </>
    )
  }

  if (error) {
    return (
      <>
        <header className="mt-4">
          <h1 className="text-3xl font-bold tracking-tight text-gray-900 sm:text-4xl">
            Marketplace
          </h1>
        </header>
        <div className="mt-6 rounded-xl border border-red-200 bg-red-50 p-6 text-center">
          <Icon name="error" size={40} className="mx-auto text-red-400" />
          <p className="mt-3 text-sm text-red-700">{error}</p>
        </div>
      </>
    )
  }

  if (offeringsList.length === 0) {
    return (
      <>
        <header className="mt-4">
          <h1 className="text-3xl font-bold tracking-tight text-gray-900 sm:text-4xl">
            Marketplace
          </h1>
        </header>
        <div className="mt-6 rounded-xl border border-dashed border-gray-300 bg-white p-12 text-center">
          <Icon name="storefront" size={48} className="mx-auto text-gray-300" />
          <h2 className="mt-4 text-lg font-semibold text-gray-900">No hay productos disponibles</h2>
          <p className="mt-2 max-w-sm mx-auto text-sm text-gray-500">
            Sé el primero en publicar productos en el Marketplace.
          </p>
        </div>
      </>
    )
  }

  return (
    <>
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
            Buscar productos
          </label>
          <input
            id="marketplace-search"
            type="text"
            value={query}
            onChange={(e) => {
              setQuery(e.target.value)
              setPage(1)
            }}
            placeholder="Buscar productos..."
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
              {paginatedProducts.map((offering) => (
                <ProductCardGrid key={offering.id} offering={offering} />
              ))}
            </div>
          ) : (
            <div className="space-y-4">
              {paginatedProducts.map((offering) => (
                <ProductCardList key={offering.id} offering={offering} />
              ))}
            </div>
          )}

          {totalPages > 1 && (
            <Pagination currentPage={page} totalPages={totalPages} onPageChange={setPage} />
          )}
        </section>
      </div>
    </>
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

function filterAndSortProducts(allProducts, query, filters, sort) {
  const normalizedQuery = query.trim().toLowerCase()

  const matching = allProducts.filter((product) => {
    if (!matchesQuery(product, normalizedQuery)) return false
    if (product.price > filters.maxPrice) return false
    return true
  })

  return sortProducts(matching, sort)
}

function matchesQuery(product, normalizedQuery) {
  if (!normalizedQuery) return true
  const searchableText = `${product.name || ''}`.toLowerCase()
  return searchableText.includes(normalizedQuery)
}

function paginateProducts(items, page) {
  const start = (page - 1) * PAGE_SIZE
  return items.slice(start, start + PAGE_SIZE)
}

function sortProducts(items, sort) {
  switch (sort) {
    case 'rating':
      return [...items].sort((a, b) => (b.rating || 0) - (a.rating || 0))
    case 'priceAsc':
      return [...items].sort((a, b) => a.price - b.price)
    case 'priceDesc':
      return [...items].sort((a, b) => b.price - a.price)
    default:
      return items
  }
}
