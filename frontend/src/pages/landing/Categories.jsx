import SectionHeading from '../../components/SectionHeading'
import CategoryPill from '../../components/CategoryPill'
import { categories } from '../../mocks/catalog'

export default function Categories() {
  return (
    <section className="bg-gray-50">
      <div className="mx-auto max-w-7xl px-4 py-16 sm:px-6 lg:px-8">
        <SectionHeading
          eyebrow="Explora"
          title="Categorías principales"
          action={
            <a href="#/marketplace" className="group inline-flex items-center gap-1 text-sm font-semibold text-brand hover:text-brand-dark">
              Ver todas
              <span aria-hidden="true" className="transition-transform group-hover:translate-x-0.5">→</span>
            </a>
          }
        />
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
          {categories.map((category) => (
            <CategoryPill key={category.id} icon={category.icon} name={category.name} count={category.count} />
          ))}
        </div>
      </div>
    </section>
  )
}