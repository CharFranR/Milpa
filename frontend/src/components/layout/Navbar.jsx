import { useEffect, useState } from 'react'
import { cn } from '../../lib/cn'
import Icon from '../ui/Icon'
import Button from '../ui/Button'

const NAV_LINKS = [
  { label: 'Inicio', href: '#inicio' },
  { label: 'Marketplace', href: '#/marketplace' },
  { label: 'Cómo funciona', href: '#como-funciona' },
]

export default function Navbar() {
  const [scrolled, setScrolled] = useState(false)
  const [open, setOpen] = useState(false)

  const transparent = !scrolled && !open

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 8)
    onScroll()
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])

  const textColor = transparent ? 'text-white' : 'text-gray-900'

  return (
    <header
      className={cn(
        'sticky top-0 z-40 transition-colors duration-300',
        transparent ? 'bg-transparent' : 'border-b border-gray-100 bg-white/95 backdrop-blur',
      )}
    >
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
        <a href="#inicio" className="flex items-center gap-2" aria-label="Milpa — inicio">
          <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-brand text-white">
            <Icon name="eco" size={20} weight={600} />
          </span>
          <span className={cn('text-lg tracking-tight', textColor)}>
            <span className={cn('font-extrabold', transparent ? 'text-accent' : 'text-brand')}>
              Mil
            </span>
            <span className="font-medium">pa</span>
          </span>
        </a>

        <nav className="hidden items-center gap-1 md:flex" aria-label="Principal">
          {NAV_LINKS.map((link) => (
            <a
              key={link.label}
              href={link.href}
              className={cn(
                'rounded-full px-4 py-2 text-sm font-medium transition-colors',
                transparent
                  ? 'text-white/80 hover:bg-white/10 hover:text-white'
                  : 'text-gray-600 hover:bg-brand-soft hover:text-brand',
              )}
            >
              {link.label}
            </a>
          ))}
        </nav>

        <div className="hidden items-center gap-2 md:flex">
          <a href="#/login">
            <Button variant={transparent ? 'white' : 'ghost'}>Ingresar</Button>
          </a>
          <a href="#/register" className="w-full sm:w-auto">
            <Button variant="accent">Registrarse gratis</Button>
          </a>
        </div>

        <button
          type="button"
          className={cn(
            'inline-flex items-center justify-center rounded-xl p-2 md:hidden',
            transparent ? 'text-white' : 'text-gray-700',
          )}
          onClick={() => setOpen((v) => !v)}
          aria-expanded={open}
          aria-label={open ? 'Cerrar menú' : 'Abrir menú'}
        >
          <Icon name={open ? 'close' : 'menu'} size={26} />
        </button>
      </div>

      {open && (
        <div className="border-t border-gray-100 bg-white px-4 pb-6 pt-3 md:hidden">
          <nav className="flex flex-col" aria-label="Menú móvil">
            {NAV_LINKS.map((link) => (
              <a
                key={link.label}
                href={link.href}
                className="rounded-xl px-3 py-3 text-sm font-medium text-gray-700 hover:bg-gray-50"
              >
                {link.label}
              </a>
            ))}
          </nav>
          <div className="mt-4 flex flex-col gap-2">
            <a href="#/login" className="w-full">
              <Button variant="outline" className="w-full">
                Ingresar
              </Button>
            </a>
            <a href="#/register" className="w-full">
              <Button variant="accent" className="w-full">
                Registrarse gratis
              </Button>
            </a>
          </div>
        </div>
      )}
    </header>
  )
}