import { useEffect, useState } from 'react'
import Icon from '../ui/Icon'
import ProductCard from '../../components/product/ProductCard'
import ProductCardList from '../../components/product/ProductCardList'
import FiltersSidebar, { PRICE_LIMIT } from './FiltersSidebar'
import Pagination from './Pagination'
import EmptyResults from './EmptyResults'
import { useSearch } from '../../hooks/useSearch'
import { cn } from '../../lib/cn'

const SORT_OPTIONS = [
  { value: 'relevance', label: 'Más relevantes' },
  { value: 'price_asc', label: 'Precio: menor a mayor' },
  { value: 'price_desc', label: 'Precio: mayor a menor' },
]

const DEFAULT_FILTERS = {
  category: 'all',
  maxPrice: PRICE_LIMIT,
}

const PAGE_SIZE = 6

export default function MarketplaceCatalog() {
  const [query, setQuery] = useState('')
  const [term, setTerm] = useState('')
  const [sort, setSort] = useState('relevance')
  const [filters, setFilters] = useState(DEFAULT_FILTERS)
  const [view, setView] = useState('grid')
  const [page, setPage] = useState(1)
  const [showFilters, setShowFilters] = useState(false)

  useEffect(() => {
    const timer = setTimeout(() => setTerm(query.trim()), 350)
    return () => clearTimeout(timer)
  }, [query])

  const { results, totalHits, totalPages, loading, error } = useSearch({
    term,
    categoryId: filters.category === 'all' ? '' : filters.category,
    maxPrice: filters.maxPrice < PRICE_LIMIT ? filters.maxPrice : null,
    sort,
    page,
    pageSize: PAGE_SIZE,
  })

  const hasActiveSearch =
    Boolean(term) || filters.category !== 'all' || filters.maxPrice < PRICE_LIMIT
  const showSkeleton = loading && results.length === 0

  function updateFilters(patch) {
    setFilters((current) => ({ ...current, ...patch }))
    setPage(1)
  }

  function clearAll() {
    setQuery('')
    setTerm('')
    setFilters(DEFAULT_FILTERS)
    setPage(1)
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

  if (showSkeleton) {
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

  if (totalHits === 0 && !hasActiveSearch) {
    return (
      <>
        <header className="mt-4">
          <h1 className="text-3xl font-bold tracking-tight text-gray-900 sm:text-4xl">
            Marketplace
          </h1>
        </header>
        <div className="mt-6 rounded-xl border border-dashed border-gray-300 bg-white p-12 text-center">
          <Icon name="storefront" size={48} className="mx-auto text-gray-400" />
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
          <span className="font-semibold text-brand">{totalHits}</span> productos
          disponibles de agricultores locales
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

        <section aria-label="Resultados" className="min-w-0 flex-1" aria-busy={loading}>
          {results.length === 0 ? (
            <EmptyResults onClear={clearAll} />
          ) : (
            <div className={cn(loading && 'opacity-60')}>
              {view === 'grid' ? (
                <div className="grid gap-5 sm:grid-cols-2 xl:grid-cols-3">
                  {results.map((offering) => (
                    <ProductCard key={offering.id} offering={toCardOffering(offering)} />
                  ))}
                </div>
              ) : (
                <div className="space-y-4">
                  {results.map((offering) => (
                    <ProductCardList key={offering.id} offering={toCardOffering(offering)} />
                  ))}
                </div>
              )}
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

function toCardOffering(result) {
  return {
    ...result,
    type: result.type === 'service' ? 1 : 0,
    company_name: result.farmer_name || '',
  }
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
