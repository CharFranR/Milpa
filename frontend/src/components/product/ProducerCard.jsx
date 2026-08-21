import Icon from '../ui/Icon'
import StarRating from '../StarRating'

export default function ProducerCard({ producer }) {
  return (
    <article className="flex flex-col rounded-2xl border border-gray-100 bg-white p-6 text-center shadow-sm transition-all duration-300 hover:-translate-y-1 hover:shadow-lg">
      <div className="mx-auto flex h-20 w-20 items-center justify-center overflow-hidden rounded-full bg-gray-200">
        <div className="h-full w-full" aria-hidden="true" />
      </div>
      <h3 className="mt-4 text-lg font-semibold text-gray-900">{producer.name}</h3>
      <p className="text-sm text-gray-500">{producer.farm}</p>
      <p className="mt-1 inline-flex items-center justify-center gap-1 text-sm text-gray-500">
        <Icon name="location_on" size={15} className="text-brand" />
        {producer.city}, {producer.region}
      </p>
      <div className="mt-3 flex items-center justify-center gap-1">
        <StarRating rating={producer.rating} reviews={producer.reviews} size={14} />
      </div>
      <div className="mt-4 flex flex-wrap justify-center gap-1.5">
        {producer.specialties.slice(0, 3).map((s) => (
          <span
            key={s}
            className="rounded-full bg-brand-soft px-2.5 py-1 text-xs font-medium text-brand"
          >
            {s}
          </span>
        ))}
      </div>
      <p className="mt-4 text-xs text-gray-400">{producer.productsCount} productos activos</p>
      <a
        href="#"
        className="mt-4 inline-flex items-center justify-center gap-1 rounded-full text-sm font-semibold text-brand transition-colors hover:text-brand-dark"
      >
        Ver perfil
        <Icon name="arrow_forward" size={16} />
      </a>
    </article>
  )
}