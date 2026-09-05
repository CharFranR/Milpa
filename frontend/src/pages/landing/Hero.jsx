import { useState } from 'react'
import Icon from '../../components/ui/Icon'
import Button from '../../components/ui/Button'
import Logo from '../../components/Logo'
import { trustChips } from '../../mocks/content'
import { regions } from '../../mocks/catalog'
import fondoCampo from '../../assets/images/fondo-campo.jpeg'

export default function Hero() {
  const [query, setQuery] = useState('')
  const [region, setRegion] = useState('todas')

  function handleSearch(e) {
    e.preventDefault()
  }

  return (
    <section className="relative isolate min-h-screen overflow-hidden">
      <div
        className="absolute inset-0 -z-10 bg-cover bg-center"
        style={{ backgroundImage: `url(${fondoCampo})` }}
      />
      <div
        aria-hidden="true"
        className="absolute inset-0 -z-10 bg-gradient-to-b from-brand-dark/80 via-night/60 to-night/90"
      />

      <div className="mx-auto flex max-w-7xl flex-col items-center px-4 pt-16 pb-24 text-center sm:px-6 sm:pt-20 lg:px-8">
        <span className="inline-flex items-center gap-1.5 rounded-full border border-white/15 bg-white/5 px-3.5 py-1.5 text-xs font-semibold text-accent">
          <Logo variant="icon" className="h-5 w-5" />
          Comercio agropecuario directo
        </span>

        <h1 className="mt-6 max-w-3xl text-4xl font-bold leading-tight tracking-tight text-white sm:text-5xl lg:text-6xl">
          Del{' '}
          <span className="relative inline-block whitespace-nowrap">
            campo
            <svg
              className="absolute -bottom-2 left-0 w-full"
              viewBox="0 0 180 12"
              fill="none"
              aria-hidden="true"
              preserveAspectRatio="none"
            >
              <path
                d="M3 9C40 3 120 2 177 7"
                stroke="#D2E749"
                strokeWidth="4"
                strokeLinecap="round"
              />
            </svg>
          </span>{' '}
          a tu mesa
        </h1>

        <p className="mt-6 max-w-2xl text-base text-white/70 sm:text-lg">
          Conecta directamente con productores locales. Frutas, verduras y productos del campo
          frescos, sin intermediarios y a precio justo.
        </p>

        <form
          onSubmit={handleSearch}
          className="mt-8 w-full max-w-3xl rounded-2xl sm:rounded-full bg-white p-4 sm:p-2 shadow-xl shadow-black/20"
        >
          <div className="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-3">
            <div className="flex min-w-0 flex-1 items-center gap-2 px-3 sm:px-4">
              <Icon name="search" size={18} className="shrink-0 text-gray-400" />
              <label htmlFor="hero-search" className="sr-only">
                Buscar productos
              </label>
              <input
                id="hero-search"
                type="search"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="Café, frijol rojo, quesillo..."
                className="w-full h-10 bg-transparent text-sm text-gray-900 placeholder:text-gray-400 focus:outline-none"
              />
            </div>
            <div className="w-full sm:w-auto">
              <label htmlFor="hero-region" className="sr-only">
                Región
              </label>
              <select
                id="hero-region"
                value={region}
                onChange={(e) => setRegion(e.target.value)}
                className="w-full h-10 bg-transparent px-3 py-0 text-sm text-gray-700 focus:outline-none"
              >
                <option value="todas">Todas las regiones</option>
                {regions.slice(0, 6).map((r) => (
                  <option key={r} value={r}>
                    {r}
                  </option>
                ))}
              </select>
            </div>
            <Button type="submit" size="md" className="w-full sm:w-auto h-12 sm:h-10">
              Buscar
            </Button>
          </div>
        </form>

        <ul className="mt-8 flex flex-wrap items-center justify-center gap-x-6 gap-y-3">
          {trustChips.map((chip) => (
            <li key={chip} className="inline-flex items-center gap-1.5 text-sm text-white/80">
              <Icon name="check_circle" size={17} className="text-accent" />
              {chip}
            </li>
          ))}
        </ul>
      </div>

      <a
        href="#como-funciona"
        className="group absolute bottom-5 left-1/2 hidden -translate-x-1/2 flex-col items-center gap-1 text-xs font-medium text-white/60 transition-colors hover:text-white sm:flex"
      >
        Explorar
        <Icon
          name="keyboard_arrow_down"
          size={20}
          className="animate-bounce text-accent motion-reduce:animate-none"
        />
      </a>
    </section>
  )
}