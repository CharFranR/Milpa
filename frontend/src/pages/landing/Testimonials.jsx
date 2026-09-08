import SectionHeading from '../../components/SectionHeading'
import Avatar from '../../components/Avatar'
import Icon from '../../components/ui/Icon'
import { testimonials } from '../../mocks/content'

export default function Testimonials() {
  return (
    <section className="bg-gray-50">
      <div className="mx-auto max-w-7xl px-4 py-16 sm:px-6 lg:px-8">
        <SectionHeading eyebrow="Lo que dicen nuestros usuarios" title="Comunidad que confía" centered />
        <div className="grid gap-6 md:grid-cols-3">
          {testimonials.map((t) => (
            <figure
              key={t.name}
              className="relative flex flex-col rounded-2xl border border-gray-100 bg-white p-6 shadow-sm"
            >
              <Icon
                name="format_quote"
                size={40}
                className="absolute right-5 top-5 text-brand/10"
              />
              <div className="flex gap-0.5 text-amber-400" aria-label="5 de 5 estrellas">
                {Array.from({ length: 5 }, (_, i) => (
                  <Icon key={i} name="star" size={16} filled weight={500} />
                ))}
              </div>
              <blockquote className="mt-4 flex-1 text-sm leading-relaxed text-gray-600">
                “{t.text}”
              </blockquote>
              <figcaption className="mt-5 flex items-center gap-3">
                <Avatar initials={t.initials} name={t.name} />
                <div>
                  <p className="text-sm font-semibold text-gray-900">{t.name}</p>
                  <p className="text-xs text-gray-500">{t.role}</p>
                </div>
              </figcaption>
            </figure>
          ))}
        </div>
      </div>
    </section>
  )
}