import { useState } from 'react'
import Icon from '../ui/Icon'
import Button from '../ui/Button'
import { regions } from '../../mocks/catalog'

const USER_TYPES = [
  {
    key: 'buyer',
    label: 'Comprador',
    icon: 'shopping_basket',
    desc: 'Quiero descubrir y comprar productos frescos directamente de productores locales.',
  },
  {
    key: 'producer',
    label: 'Productor',
    icon: 'agriculture',
    desc: 'Quiero publicar mis productos y venderlos sin intermediarios.',
  },
]

export default function Register() {
  const [step, setStep] = useState(1)
  const [userType, setUserType] = useState('')
  const [form, setForm] = useState({
    name: '',
    email: '',
    phone: '',
    password: '',
    farm: '',
    region: '',
  })
  const [legal, setLegal] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const isProducer = userType === 'producer'

  function setField(key, value) {
    setForm((f) => ({ ...f, [key]: value }))
  }

  function goToStep2() {
    if (!userType) return
    setError('')
    setStep(2)
  }

  function handleSubmit(e) {
    e.preventDefault()
    if (!form.name.trim() || !form.email.trim() || !form.phone.trim() || !form.password) {
      setError('Por favor completa todos los campos.')
      return
    }
    if (isProducer && (!form.farm.trim() || !form.region)) {
      setError('Por favor completa los datos de tu finca.')
      return
    }
    if (!legal) {
      setError('Debes aceptar los Términos de uso y la Política de privacidad.')
      return
    }
    setError('')
    setLoading(true)
    setTimeout(() => setLoading(false), 1200)
  }

  return (
    <div className="m-0 flex min-h-screen items-center justify-center bg-gray-50 p-4 sm:p-8 md:m-0 md:min-h-screen md:justify-normal md:items-stretch">
      <aside className="hidden bg-brand text-white md:relative md:flex md:w-1/2 md:flex-col md:justify-between md:overflow-hidden">
        <div className="p-8">
          <a href="#/" className="flex items-center gap-2" aria-label="Milpa — inicio">
            <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-white/20">
              <Icon name="eco" size={20} weight={600} className="text-accent" />
            </span>
            <span className="text-lg font-medium">
              Mil<span className="font-extrabold">pa</span>
            </span>
          </a>
        </div>

        <div className="p-10">
          <div className="flex h-20 w-20 items-center justify-center rounded-2xl bg-white/15">
            <Icon name="agriculture" size={40} />
          </div>
          <h2 className="mt-5 text-2xl font-bold">Únete a la red Milpa</h2>
          <p className="mt-3 max-w-xs text-sm opacity-85">
            Crea tu cuenta en dos pasos y empieza a comprar o vender productos frescos del campo.
          </p>
          <ul className="mt-6 space-y-3 text-sm">
            <li className="flex items-center gap-2.5">
              <Icon name="verified_user" size={18} />
              Compra o vende sin intermediarios
            </li>
            <li className="flex items-center gap-2.5">
              <Icon name="handshake" size={18} />
              Coordinación directa entre usuarios
            </li>
            <li className="flex items-center gap-2.5">
              <Icon name="storefront" size={18} />
              Perfil público para productores
            </li>
          </ul>
        </div>

        <div className="p-8 text-center text-xs opacity-60">
          © 2025 Milpa. Todos los derechos reservados.
        </div>
      </aside>

      <main className="w-full max-w-md space-y-6">
        <header className="flex justify-between">
          <a href="#/" className="flex items-center gap-2" aria-label="Milpa — inicio">
            <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-brand">
              <Icon name="eco" size={20} weight={600} className="text-white" />
            </span>
            <span className="text-lg font-medium text-gray-900">
              Mil<span className="font-extrabold text-brand">pa</span>
            </span>
          </a>
          <p className="text-sm text-gray-500">
            ¿Ya tienes cuenta?{' '}
            <a href="#/login" className="font-semibold text-brand hover:text-brand-dark">
              Ingresar
            </a>
          </p>
        </header>

        {/* Indicador de pasos */}
        <ol className="flex items-center gap-3">
          {[1, 2].map((n) => (
            <li key={n} className="flex flex-1 items-center gap-3">
              <span
                className={`flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-xs font-bold ${
                  step >= n ? 'bg-brand text-white' : 'bg-gray-200 text-gray-500'
                }`}
              >
                {step > n ? <Icon name="check" size={16} /> : n}
              </span>
              <span
                className={`text-xs font-semibold ${
                  step >= n ? 'text-brand' : 'text-gray-400'
                }`}
              >
                {n === 1 ? 'Tipo de cuenta' : 'Tus datos'}
              </span>
              {n === 1 && <span className="h-px flex-1 bg-gray-300" />}
            </li>
          ))}
        </ol>

        {step === 1 && (
          <section>
            <h1 className="text-xl font-bold text-gray-900">¿Cómo vas a usar Milpa?</h1>
            <p className="mt-1 text-sm text-gray-500">
              Elige el tipo de cuenta que mejor se adapte a ti.
            </p>

            <div className="mt-5 grid gap-3">
              {USER_TYPES.map((t) => (
                <button
                  key={t.key}
                  type="button"
                  onClick={() => setUserType(t.key)}
                  className={`flex items-start gap-3 rounded-xl border p-4 text-left transition ${
                    userType === t.key
                      ? 'border-brand bg-brand-soft ring-1 ring-brand'
                      : 'border-gray-300 bg-white hover:border-gray-400'
                  }`}
                >
                  <span
                    className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-lg ${
                      userType === t.key ? 'bg-brand text-white' : 'bg-gray-100 text-gray-600'
                    }`}
                  >
                    <Icon name={t.icon} size={22} />
                  </span>
                  <span className="min-w-0">
                    <span className="block text-sm font-bold text-gray-900">{t.label}</span>
                    <span className="mt-0.5 block text-xs leading-relaxed text-gray-500">
                      {t.desc}
                    </span>
                  </span>
                  {userType === t.key && (
                    <Icon name="check_circle" size={22} className="ml-auto shrink-0 text-brand" />
                  )}
                </button>
              ))}
            </div>

            {error && (
              <p className="mt-4 rounded-lg bg-red-50 px-3 py-2 text-xs text-red-700">{error}</p>
            )}

            <Button
              type="button"
              variant="primary"
              size="lg"
              disabled={!userType}
              onClick={goToStep2}
              className="mt-5 w-full"
            >
              Continuar
            </Button>
          </section>
        )}

        {step === 2 && (
          <form onSubmit={handleSubmit} className="space-y-4">
            <h1 className="text-xl font-bold text-gray-900">
              Crea tu cuenta de {isProducer ? 'productor' : 'comprador'}
            </h1>

            <div>
              <label htmlFor="reg-name" className="text-xs font-semibold text-gray-600">
                Nombre completo
              </label>
              <input
                id="reg-name"
                type="text"
                required
                value={form.name}
                onChange={(e) => setField('name', e.target.value)}
                placeholder="Tu nombre"
                className="mt-1.5 w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2.5 text-sm text-gray-900 placeholder-gray-400 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
              />
            </div>

            <div>
              <label htmlFor="reg-email" className="text-xs font-semibold text-gray-600">
                Correo electrónico
              </label>
              <input
                id="reg-email"
                type="email"
                required
                value={form.email}
                onChange={(e) => setField('email', e.target.value)}
                placeholder="tu@correo.com"
                className="mt-1.5 w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2.5 text-sm text-gray-900 placeholder-gray-400 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
              />
            </div>

            <div>
              <label htmlFor="reg-phone" className="text-xs font-semibold text-gray-600">
                Teléfono
              </label>
              <input
                id="reg-phone"
                type="tel"
                required
                value={form.phone}
                onChange={(e) => setField('phone', e.target.value)}
                placeholder="+504 0000-0000"
                className="mt-1.5 w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2.5 text-sm text-gray-900 placeholder-gray-400 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
              />
            </div>

            <div>
              <label htmlFor="reg-password" className="text-xs font-semibold text-gray-600">
                Contraseña
              </label>
              <input
                id="reg-password"
                type="password"
                required
                minLength={8}
                value={form.password}
                onChange={(e) => setField('password', e.target.value)}
                placeholder="Mínimo 8 caracteres"
                className="mt-1.5 w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2.5 text-sm text-gray-900 placeholder-gray-400 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
              />
            </div>

            {isProducer && (
              <>
                <div>
                  <label htmlFor="reg-farm" className="text-xs font-semibold text-gray-600">
                    Nombre de tu finca o negocio
                  </label>
                  <input
                    id="reg-farm"
                    type="text"
                    value={form.farm}
                    onChange={(e) => setField('farm', e.target.value)}
                    placeholder="Finca La Esperanza"
                    className="mt-1.5 w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2.5 text-sm text-gray-900 placeholder-gray-400 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
                  />
                </div>
                <div>
                  <label htmlFor="reg-region" className="text-xs font-semibold text-gray-600">
                    Departamento
                  </label>
                  <select
                    id="reg-region"
                    value={form.region}
                    onChange={(e) => setField('region', e.target.value)}
                    className="mt-1.5 w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-2.5 text-sm text-gray-900 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
                  >
                    <option value="">Selecciona un departamento</option>
                    {regions.map((r) => (
                      <option key={r} value={r}>
                        {r}
                      </option>
                    ))}
                  </select>
                </div>
              </>
            )}

            <label className="flex items-start gap-2.5">
              <input
                type="checkbox"
                checked={legal}
                onChange={(e) => setLegal(e.target.checked)}
                className="mt-0.5 h-4 w-4 accent-[#075809]"
              />
              <span className="text-xs leading-relaxed text-gray-500">
                Al registrarte aceptas nuestros{' '}
                <a href="#" className="font-semibold text-brand hover:underline">
                  Términos de uso
                </a>{' '}
                y{' '}
                <a href="#" className="font-semibold text-brand hover:underline">
                  Política de privacidad
                </a>
                .
              </span>
            </label>

            {error && (
              <p className="rounded-lg bg-red-50 px-3 py-2 text-xs text-red-700">{error}</p>
            )}

            <div className="flex gap-3 pt-1">
              <Button
                type="button"
                variant="outline"
                size="lg"
                onClick={() => setStep(1)}
                className="w-1/3"
              >
                Volver
              </Button>
              <Button type="submit" variant="primary" size="lg" disabled={loading} className="flex-1">
                {loading ? (
                  <>
                    <Icon name="progress_activity" size={18} className="animate-spin" />
                    Creando cuenta...
                  </>
                ) : (
                  'Crear cuenta'
                )}
              </Button>
            </div>
          </form>
        )}
      </main>
    </div>
  )
}
