import Icon from '../../components/ui/Icon'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'
import { producerBusiness } from '../../mocks/producer'

export default function ProducerBusiness() {
  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold text-gray-900">Mi negocio</h1>

      <div className="grid gap-6 lg:grid-cols-2">
        <section aria-label="Información de la finca" className="rounded-xl border border-gray-100 bg-white p-6">
          <div className="flex items-start gap-4">
            <img
              src={producerBusiness.farmPhoto}
              alt=""
              className="h-16 w-16 shrink-0 rounded-full bg-gray-100 object-cover"
            />
            <div className="min-w-0 flex-1">
              <h2 className="text-xl font-bold text-gray-900">{producerBusiness.farmName}</h2>
              <p className="mt-0.5 text-gray-500">{producerBusiness.location}</p>
              <p className="mt-1 text-sm text-gray-500">Productora desde {producerBusiness.since}</p>
            </div>
          </div>
          <dl className="mt-6 space-y-4">
            <div className="flex items-center gap-3">
              <Icon name="mail" size={18} className="text-gray-400 shrink-0" />
              <div>
                <dt className="text-xs font-semibold text-gray-500">Correo</dt>
                <dd className="text-gray-900">{producerBusiness.contact.email}</dd>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <Icon name="phone" size={18} className="text-gray-400 shrink-0" />
              <div>
                <dt className="text-xs font-semibold text-gray-500">Teléfono</dt>
                <dd className="text-gray-900">{producerBusiness.contact.phone}</dd>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <Icon name="schedule" size={18} className="text-gray-400 shrink-0" />
              <div>
                <dt className="text-xs font-semibold text-gray-500">Horario</dt>
                <dd className="text-gray-900">{producerBusiness.contact.schedule}</dd>
              </div>
            </div>
          </dl>
          <Button
            type="button"
            variant="outline"
            className="mt-6 w-full"
          >
            Editar información
          </Button>
        </section>

        <section aria-label="Especialidades y redes" className="rounded-xl border border-gray-100 bg-white p-6">
          <h2 className="text-lg font-semibold text-gray-900">Especialidades</h2>
          <div className="mt-3 flex flex-wrap gap-2">
            {producerBusiness.specialties.map((s) => (
              <Badge key={s} tone="brand-soft" className="text-xs">
                {s}
              </Badge>
            ))}
          </div>

          <h2 className="mt-6 text-lg font-semibold text-gray-900">Redes sociales</h2>
          <div className="mt-3 space-y-3">
            <a
              href={`https://instagram.com/${producerBusiness.social.instagram.replace('@', '')}`}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-3 text-sm text-gray-600 hover:text-brand"
            >
              <span className="inline-flex h-8 w-8 items-center justify-center rounded-lg bg-brand-soft text-brand">
                <Icon name="camera_alt" size={18} />
              </span>
              <span>Instagram: {producerBusiness.social.instagram}</span>
            </a>
            <a
              href={`https://facebook.com/${producerBusiness.social.facebook.replace(' ', '')}`}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-3 text-sm text-gray-600 hover:text-brand"
            >
              <span className="inline-flex h-8 w-8 items-center justify-center rounded-lg bg-brand-soft text-brand">
                <Icon name="facebook" size={18} />
              </span>
              <span>Facebook: {producerBusiness.social.facebook}</span>
            </a>
          </div>
        </section>
      </div>
    </div>
  )
}