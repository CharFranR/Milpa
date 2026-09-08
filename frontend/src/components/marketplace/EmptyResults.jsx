import Icon from '../ui/Icon'
import Button from '../ui/Button'

export default function EmptyResults({ onClear }) {
  return (
    <div className="flex flex-col items-center rounded-2xl border border-dashed border-gray-200 bg-white px-6 py-16 text-center">
      <span className="flex h-16 w-16 items-center justify-center rounded-full bg-brand-soft">
        <Icon name="search_off" size={32} className="text-brand" />
      </span>
      <h3 className="mt-4 text-lg font-bold text-gray-900">No encontramos productos</h3>
      <p className="mt-1 text-sm text-gray-500">
        Intenta con otros filtros o términos de búsqueda.
      </p>
      <Button type="button" variant="primary" onClick={onClear} className="mt-6">
        Limpiar filtros
      </Button>
    </div>
  )
}
