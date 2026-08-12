import { useState } from 'react'
import SectionHeading from '../../components/SectionHeading'
import Icon from '../../components/ui/Icon'
import { faqs } from '../../mocks/content'

export default function Faq() {
  const [open, setOpen] = useState(0)

  return (
    <section className="bg-white">
      <div className="mx-auto max-w-3xl px-4 py-16 sm:px-6 lg:px-8">
        <SectionHeading eyebrow="Preguntas frecuentes" title="¿Tienes dudas?" centered />
        <div className="space-y-3">
          {faqs.map((faq, index) => {
            const isOpen = open === index
            return (
              <div
                key={faq.question}
                className={`overflow-hidden rounded-2xl border transition-colors ${
                  isOpen ? 'border-brand/20 bg-brand-soft' : 'border-gray-100 bg-gray-50'
                }`}
              >
                <h3>
                  <button
                    type="button"
                    onClick={() => setOpen(isOpen ? -1 : index)}
                    aria-expanded={isOpen}
                    aria-controls={`faq-panel-${index}`}
                    className="flex w-full items-center justify-between gap-4 px-5 py-4 text-left"
                  >
                    <span className="text-sm font-semibold text-gray-900 sm:text-base">
                      {faq.question}
                    </span>
                    <Icon
                      name="expand_more"
                      size={22}
                      className={`shrink-0 text-brand transition-transform duration-300 ${
                        isOpen ? 'rotate-180' : ''
                      }`}
                    />
                  </button>
                </h3>
                {isOpen && (
                  <div
                    id={`faq-panel-${index}`}
                    className="px-5 pb-5 text-sm leading-relaxed text-gray-600"
                  >
                    {faq.answer}
                  </div>
                )}
              </div>
            )
          })}
        </div>
      </div>
    </section>
  )
}