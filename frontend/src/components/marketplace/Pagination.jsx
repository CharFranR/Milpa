import Icon from '../ui/Icon'
import { cn } from '../../lib/cn'

const BASE_BUTTON =
  'inline-flex h-9 min-w-9 items-center justify-center rounded-lg border border-gray-200 px-2 text-sm font-semibold transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand disabled:cursor-not-allowed disabled:opacity-40'

export default function Pagination({ currentPage, totalPages, onPageChange }) {
  const pages = Array.from({ length: totalPages }, (_, index) => index + 1)

  return (
    <nav aria-label="Paginación de resultados" className="mt-8 flex items-center justify-center gap-1.5">
      <button
        type="button"
        onClick={() => onPageChange(currentPage - 1)}
        disabled={currentPage === 1}
        aria-label="Página anterior"
        className={cn(BASE_BUTTON, 'bg-white text-gray-600 hover:bg-brand-soft hover:text-brand')}
      >
        <Icon name="chevron_left" size={18} />
      </button>

      {pages.map((pageNumber) => (
        <button
          key={pageNumber}
          type="button"
          onClick={() => onPageChange(pageNumber)}
          aria-current={pageNumber === currentPage ? 'page' : undefined}
          className={cn(
            BASE_BUTTON,
            pageNumber === currentPage
              ? 'border-brand bg-brand text-white'
              : 'bg-white text-gray-700 hover:bg-brand-soft hover:text-brand',
          )}
        >
          {pageNumber}
        </button>
      ))}

      <button
        type="button"
        onClick={() => onPageChange(currentPage + 1)}
        disabled={currentPage === totalPages}
        aria-label="Página siguiente"
        className={cn(BASE_BUTTON, 'bg-white text-gray-600 hover:bg-brand-soft hover:text-brand')}
      >
        <Icon name="chevron_right" size={18} />
      </button>
    </nav>
  )
}
