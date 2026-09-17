import Button from '../../components/ui/Button'
import Logo from '../../components/Logo'
import fondoCampo from '../../assets/images/fondo-campo.jpeg'

export default function CtaBanner() {
  return (
    <section
      className="relative isolate overflow-hidden bg-cover bg-center"
      style={{ backgroundImage: `url(${fondoCampo})` }}
    >
      <div
        aria-hidden="true"
        className="absolute inset-0 -z-10 bg-gradient-to-b from-night/90 via-night/70 to-night/95"
      />
      <div className="mx-auto flex max-w-4xl flex-col items-center px-4 py-20 text-center sm:px-6">
        <span className="inline-flex items-center gap-1.5 rounded-full bg-accent px-3.5 py-1.5 text-xs font-bold text-night">
          <Logo variant="icon" className="h-5 w-5" />
          Empieza hoy — es gratis
        </span>
        <h2 className="mt-5 text-3xl font-bold tracking-tight text-white sm:text-4xl">
          El campo está esperando por ti
        </h2>
        <p className="mt-4 max-w-2xl text-base text-white/70">
          Únete a más de 2.400 productores y miles de compradores que ya disfrutan del comercio
          justo y directo en Nicaragua.
        </p>
        <div className="mt-8 flex flex-col gap-3 sm:flex-row">
          <a href="#" className="w-full sm:w-auto">
            <Button variant="accent" size="lg" className="w-full">
              Soy comprador — quiero productos frescos
            </Button>
          </a>
          <a href="#" className="w-full sm:w-auto">
            <Button variant="white" size="lg" className="w-full">
              Soy productor — quiero vender mis productos
            </Button>
          </a>
        </div>
      </div>
    </section>
  )
}