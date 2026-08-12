import SectionHeading from '../../components/SectionHeading'
import ProducerCard from '../../components/product/ProducerCard'
import { producers } from '../../mocks/catalog'

export default function FeaturedProducers() {
  const featured = producers.slice(0, 3)

  return (
    <section className="bg-white">
      <div className="mx-auto max-w-7xl px-4 py-16 sm:px-6 lg:px-8">
        <SectionHeading
          eyebrow="Conoce a quienes cultivan"
          title="Productores destacados"
          action={
            <a href="#" className="group inline-flex items-center gap-1 text-sm font-semibold text-brand hover:text-brand-dark">
              Ver todos
              <span aria-hidden="true" className="transition-transform group-hover:translate-x-0.5">→</span>
            </a>
          }
        />
        <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {featured.map((producer) => (
            <ProducerCard key={producer.id} producer={producer} />
          ))}
        </div>
      </div>
    </section>
  )
}