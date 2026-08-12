import SectionHeading from '../../components/SectionHeading'
import Icon from '../../components/ui/Icon'
import { howItWorks } from '../../mocks/content'

export default function HowItWorks() {
  return (
    <section id="como-funciona" className="scroll-mt-20 bg-white">
      <div className="mx-auto max-w-7xl px-4 py-16 sm:px-6 lg:px-8">
        <SectionHeading eyebrow="Simple y transparente" title="Así funciona EcoMercado" centered />
        <ol className="relative grid gap-6 md:grid-cols-3">
          {howItWorks.map((step, index) => (
            <li key={step.title} className="relative rounded-2xl border border-gray-100 bg-gray-50 p-7">
              <span
                aria-hidden="true"
                className="absolute right-5 top-4 text-5xl font-bold text-brand/10"
              >
                {String(index + 1).padStart(2, '0')}
              </span>
              <span className="inline-flex h-12 w-12 items-center justify-center rounded-2xl bg-brand text-white">
                <Icon name={step.icon} size={24} />
              </span>
              <h3 className="mt-4 text-lg font-bold text-gray-900">{step.title}</h3>
              <p className="mt-2 text-sm leading-relaxed text-gray-600">{step.description}</p>
            </li>
          ))}
        </ol>
      </div>
    </section>
  )
}