import { useState } from 'react'
import Icon from '../../components/ui/Icon'
import Button from '../../components/ui/Button'
import { setSessionRole } from '../../lib/session'

const ROLES = [
  {
    key: 'buyer',
    label: 'Comprador',
    icon: 'shopping_basket',
    panel: 'bg-brand text-white',
    submit: 'primary',
    title: 'El campo más cerca que nunca',
    desc: 'Explora productos frescos de productores locales y coordina tu compra directamente, sin intermediarios.',
    bullets: [
      { icon: 'verified_user', text: 'Productores verificados' },
      { icon: 'local_shipping', text: 'Coordinación directa' },
      { icon: 'price_check', text: 'Precio justo sin intermediarios' },
    ],
  },
  {
    key: 'producer',
    label: 'Productor',
    icon: 'agriculture',
    panel: 'bg-accent text-night',
    submit: 'dark',
    title: 'Conecta tu cosecha con el mundo',
    desc: 'Publica tus productos y llega a compradores que valoran tu trabajo y la frescura del campo.',
    bullets: [
      { icon: 'bar_chart', text: 'Estadísticas de tu negocio' },
      { icon: 'inbox', text: 'Gestión de solicitudes de compra' },
      { icon: 'storefront', text: 'Perfil público de tu finca' },
    ],
  },
]

const GOOGLE_SVG = (
  <svg width="18" height="18" viewBox="0 0 44 44" aria-hidden="true" fill="none">
    <path
      d="M22.04 18.182v8.046h6.957c-.306 1.732-1.23 3.15-2.62 4.028l-.023-.018 4.934 3.85 3.882 2.982c2.398-2.2 3.794-5.426 3.794-8.994 0-2.318-.43-4.527-1.23-6.529l.047-.027z"
      fill="#FBC02D"
    />
    <path
      d="M12.264 27.857A11.96 11.96 0 0 0 22.04 34.53c1.479-.124 2.926-.38 4.318-.875l-.046-.027a13.02 13.02 0 0 0 2.236-.668l-.037-.023a11.89 11.89 0 0 0 3.053-4.725l-.04-.02v-.002a11.98 11.98 0 0 1-17.646-8.192c0-2.08.556-4.07 1.54-5.757l-.013-.01A12.003 12.003 0 0 0 12.264 27.857z"
      fill="#1E8E3E"
    />
    <path
      d="M8.76 23.614c-.316-1.58-.486-3.19-.486-4.814 0-3.758 2.09-7.05 5.276-8.933-.503 1.02-.786 2.134-.786 3.318 0 4.878 2.13 8.986 5.514 11.04l-.006.005A12.046 12.046 0 0 0 12.264 27.857z"
      fill="#1976D2"
    />
  </svg>
)

export default function Login() {
  const [role, setRole] = useState('buyer')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [showPass, setShowPass] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const current = ROLES.find((r) => r.key === role)

  function handleSubmit(e) {
    e.preventDefault()
    if (!email.trim() || !password) {
      setError('Por favor completa todos los campos.')
      return
    }
    setError('')
    setLoading(true)
    setTimeout(() => {
      setLoading(false)
      setSessionRole(current.key)
      if (current.key === 'buyer') {
        window.location.hash = '#/dashboard'
      }
    }, 1000)
  }

  return (
    <div className="m-0 flex min-h-screen items-center justify-center bg-gray-50 p-4 sm:p-8 md:m-0 md:mt-0 md:min-h-screen md:justify-normal md:items-stretch md:gap-8">
      <aside
        className={`hidden md:flex md:w-1/2 md:flex-col md:justify-between ${current.panel} relative overflow-hidden md:rounded-3xl`}
      >
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
            <Icon name={current.icon} size={40} />
          </div>
          <h2 className="mt-5 text-2xl font-bold">{current.title}</h2>
          <p className="mt-3 max-w-xs text-sm opacity-85">{current.desc}</p>
          <ul className="mt-6 space-y-3">
            {current.bullets.map((b) => (
              <li key={b.text} className="flex items-center gap-2.5 text-sm">
                <Icon name={b.icon} size={18} />
                {b.text}
              </li>
            ))}
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
          <a href="#/register" className="text-sm font-semibold text-brand hover:text-brand-dark">
            Regístrate
          </a>
        </header>

        <div>
          <p className="mb-2 text-xs font-semibold uppercase tracking-wider text-gray-500">
            Ingresar como
          </p>
          <div className="grid grid-cols-3 gap-2">
            {ROLES.map((r) => (
              <button
                key={r.key}
                type="button"
                onClick={() => setRole(r.key)}
                className={`rounded-xl px-3 py-2 text-xs font-semibold transition ${
                  role === r.key
                    ? `${r.panel} ring-2 ring-brand`
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
              >
                {r.label}
              </button>
            ))}
          </div>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <h1 className="text-xl font-bold text-gray-900">Bienvenido de vuelta</h1>

          <div>
            <label htmlFor="login-email" className="text-xs font-semibold text-gray-600">
              Correo electrónico
            </label>
            <div className="relative mt-1.5">
              <Icon
                name="mail"
                size={18}
                className="absolute left-3 top-3 text-gray-400"
              />
              <input
                id="login-email"
                type="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="tu@correo.com"
                className="w-full rounded-lg border border-gray-300 bg-gray-50 py-2.5 pl-10 text-sm text-gray-900 placeholder-gray-400 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
              />
            </div>
          </div>

          <div>
            <label htmlFor="login-password" className="text-xs font-semibold text-gray-600">
              Contraseña
            </label>
            <div className="relative mt-1.5">
              <Icon
                name="lock"
                size={18}
                className="absolute left-3 top-3 text-gray-400"
              />
              <input
                id="login-password"
                type={showPass ? 'text' : 'password'}
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Tu contraseña"
                className="w-full rounded-lg border border-gray-300 bg-gray-50 py-2.5 pl-10 pr-10 text-sm text-gray-900 placeholder-gray-400 focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand"
              />
              <button
                type="button"
                onClick={() => setShowPass((v) => !v)}
                aria-label={showPass ? 'Ocultar contraseña' : 'Mostrar contraseña'}
                className="absolute right-2 top-1/2 -translate-y-1/2 rounded-full p-1 text-gray-400 hover:text-gray-600"
              >
                <Icon name={showPass ? 'visibility_off' : 'visibility'} size={20} />
              </button>
            </div>
            <a
              href="#"
              className="mt-1 block text-right text-xs font-semibold text-brand hover:underline"
            >
              ¿Olvidaste tu contraseña?
            </a>
          </div>

          {error && (
            <p className="rounded-lg bg-red-50 px-3 py-2 text-xs text-red-700">{error}</p>
          )}

          <Button
            type="submit"
            variant={current.submit}
            size="lg"
            disabled={loading}
            className="w-full"
          >
            {loading ? (
              <>
                <Icon name="progress_activity" size={18} className="animate-spin" />
                Verificando...
              </>
            ) : (
              `Ingresar como ${current.label}`
            )}
          </Button>

          <div className="my-4 flex items-center gap-2 text-xs text-gray-400">
            <span className="h-px flex-1 bg-gray-300" />o continúa con
            <span className="h-px flex-1 bg-gray-300" />
          </div>

          <Button
            type="button"
            variant="outline"
            size="lg"
            className="w-full"
            icon={<span className="flex h-5 w-5 items-center justify-center">{GOOGLE_SVG}</span>}
          >
            Continuar con Google
          </Button>
        </form>
      </main>
    </div>
  )
}