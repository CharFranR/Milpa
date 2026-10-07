import { useState } from 'react'
import Icon from '../ui/Icon'
import Logo from '../../components/Logo'
import { Link } from 'react-router-dom'

const PLATFORM_LINKS = [
  { label: 'Para compradores' },
  { label: 'Para agricultores' },
  { label: 'Registrarse', to: '/register' },
  { label: 'Iniciar sesión', to: '/login' },
]

const COMPANY_LINKS = ['Sobre nosotros', 'Contacto']

export default function Footer() {
  const [email, setEmail] = useState('')
  const [sent, setSent] = useState(false)

  function handleSubscribe(e) {
    e.preventDefault()
    if (!email.trim()) return
    setEmail('')
    setSent(true)
  }

  return (
    <footer className="bg-night text-gray-300">
      <div className="mx-auto max-w-7xl px-4 py-14 sm:px-6 lg:px-8">
        <div className="grid gap-10 md:grid-cols-2 lg:grid-cols-4">
          <div>
            <p className="flex items-center gap-2 text-white">
              <Logo className="h-9 w-auto" />
            </p>
            <p className="mt-4 max-w-xs text-sm text-gray-400">
              Conectando el campo con la ciudad. Productos frescos, directamente del agricultor a tu
              mesa.
            </p>
          </div>

          <nav aria-label="Plataforma">
            <h3 className="text-xs font-semibold uppercase tracking-widest text-white">Plataforma</h3>
            <ul className="mt-4 space-y-2.5">
              {PLATFORM_LINKS.map((link) => (
                <li key={link.label}>
                  {link.to ? (
                    <Link to={link.to} className="text-sm text-gray-400 transition-colors hover:text-white">
                      {link.label}
                    </Link>
                  ) : (
                    <span className="text-sm text-gray-400">{link.label}</span>
                  )}
                </li>
              ))}
            </ul>
          </nav>

          <nav aria-label="Empresa">
            <h3 className="text-xs font-semibold uppercase tracking-widest text-white">Empresa</h3>
            <ul className="mt-4 space-y-2.5">
              {COMPANY_LINKS.map((label) => (
                <li key={label}>
                  <a href="#" className="text-sm text-gray-400 transition-colors hover:text-white">
                    {label}
                  </a>
                </li>
              ))}
            </ul>
          </nav>

          <div>
            <h3 className="text-xs font-semibold uppercase tracking-widest text-white">Novedades</h3>
            <p className="mt-4 text-sm text-gray-400">
              Recibe alertas de nuevos productos y productores en tu región.
            </p>
            <form onSubmit={handleSubscribe} className="mt-4">
              <label htmlFor="newsletter-email" className="sr-only">
                Correo electrónico para novedades
              </label>
              <div className="flex rounded-full bg-white/5 p-1 focus-within:ring-2 focus-within:ring-brand">
                <input
                  id="newsletter-email"
                  type="email"
                  required
                  value={email}
                  onChange={(e) => {
                    setEmail(e.target.value)
                    setSent(false)
                  }}
                  placeholder="tu@correo.com"
                  className="w-full bg-transparent px-3 text-sm text-white placeholder:text-gray-500 focus:outline-none focus-visible:outline-2 focus-visible:outline-white"
                />
                <button
                  type="submit"
                  aria-label="Suscribirse a novedades"
                  className="inline-flex h-10 w-11 shrink-0 items-center justify-center rounded-full bg-accent text-night transition-colors hover:brightness-95"
                >
                  <Icon name="send" size={18} />
                </button>
              </div>
              {sent && (
                <p className="mt-2 text-xs text-accent" role="status">
                  ¡Listo! Te avisaremos cuando haya novedades.
                </p>
              )}
            </form>
          </div>
        </div>

        <div className="mt-12 flex flex-col items-center justify-between gap-3 border-t border-white/10 pt-6 sm:flex-row">
          <p className="text-sm text-gray-500">© 2026 Milpa. Todos los derechos reservados.</p>
          <ul className="flex gap-5">
            {['Privacidad', 'Términos', 'Cookies'].map((label) => (
              <li key={label}>
                <a href="#" className="text-sm text-gray-500 transition-colors hover:text-white">
                  {label}
                </a>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </footer>
  )
}