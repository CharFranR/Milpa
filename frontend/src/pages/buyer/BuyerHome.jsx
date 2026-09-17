import Icon from '../../components/ui/Icon'
import StatCard from '../../components/StatCard'
import ProductImage from '../../components/product/ProductImage'
import { formatPrice } from '../../lib/format'
import { getUser } from '../../lib/session'
import { getDisplayName } from '../../lib/user'
import { productById, producerById } from '../../mocks/catalog'
import { homeStats, quickActions, recentActivity, recommendedProductIds } from '../../mocks/buyer'

const ACTIVITY_TONES = {
  red: 'bg-red-100 text-red-600',
  blue: 'bg-blue-100 text-blue-600',
  brand: 'bg-brand-soft text-brand',
}

export default function BuyerHome({ onGoToTab }) {
  const user = getUser()
  const displayName = getDisplayName(user)

  return (
    <div className="space-y-8">
      <header>
        <h1 className="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">
          Buen día, {displayName} 👋
        </h1>
        <p className="mt-1 text-sm text-gray-500">
          Aquí tienes un resumen de tu actividad reciente.
        </p>
      </header>

      <section aria-label="Resumen de actividad">
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
          {homeStats.map((stat) => (
            <StatCard key={stat.label} icon={stat.icon} value={stat.value} label={stat.label} tone={stat.tone} />
          ))}
        </div>
      </section>

      <section aria-label="Accesos rápidos">
        <h2 className="text-base font-bold text-gray-900">Accesos rápidos</h2>
        <div className="mt-3 grid grid-cols-2 gap-4 xl:grid-cols-4">
          {quickActions.map((action) => {
            const className =
              'flex flex-col items-start gap-3 rounded-2xl border border-gray-100 bg-white p-4 text-left transition-colors hover:border-brand/40 hover:bg-brand-soft/50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand'
            const iconTile = (
              <>
                <span className="inline-flex h-10 w-10 items-center justify-center rounded-xl bg-brand-soft text-brand">
                  <Icon name={action.icon} size={20} />
                </span>
                <span className="text-sm font-semibold text-gray-800">{action.label}</span>
              </>
            )
            return action.tab ? (
              <button key={action.label} type="button" onClick={() => onGoToTab(action.tab)} className={className}>
                {iconTile}
              </button>
            ) : (
              <a key={action.label} href={action.href} className={className}>
                {iconTile}
              </a>
            )
          })}
        </div>
      </section>

      <section aria-label="Productos recomendados">
        <div className="flex items-center justify-between">
          <h2 className="text-base font-bold text-gray-900">Recomendados para ti</h2>
          <button
            type="button"
            onClick={() => onGoToTab('marketplace')}
            className="group inline-flex items-center gap-1 text-sm font-semibold text-brand hover:text-brand-dark focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand"
          >
            Ver todos
            <span aria-hidden="true" className="transition-transform group-hover:translate-x-0.5">
              →
            </span>
          </button>
        </div>
        <div className="mt-3 grid grid-cols-2 gap-4 xl:grid-cols-4">
          {recommendedProductIds.map((productId) => {
            const product = productById(productId)
            if (!product) return null
            const producer = producerById(product.producerId)
            return (
              <article
                key={productId}
                className="group overflow-hidden rounded-2xl border border-gray-100 bg-white transition-all duration-300 hover:-translate-y-0.5 hover:shadow-md motion-reduce:transition-none motion-reduce:hover:translate-y-0"
              >
                <div className="aspect-[4/3]">
                  <ProductImage productId={product.id} name={product.name} />
                </div>
                <div className="p-3">
                  <h3 className="truncate text-sm font-semibold text-gray-900">{product.name}</h3>
                  <p className="mt-0.5 truncate text-xs text-gray-500">{producer.city}, {producer.region}</p>
                  <p className="mt-1.5 text-sm font-bold text-brand">
                    {formatPrice(product.price)}
                    <span className="ml-1 text-xs font-medium text-gray-400">/ {product.unit}</span>
                  </p>
                </div>
              </article>
            )
          })}
        </div>
      </section>

      <section aria-label="Actividad reciente">
        <h2 className="text-base font-bold text-gray-900">Actividad reciente</h2>
        <ul className="mt-3 divide-y divide-gray-100 rounded-2xl border border-gray-100 bg-white">
          {recentActivity.map((item) => (
            <li key={item.text} className="flex items-center gap-3 px-5 py-4">
              <span
                className={`inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-xl ${ACTIVITY_TONES[item.tone]}`}
              >
                <Icon name={item.icon} size={18} />
              </span>
              <p className="min-w-0 flex-1 truncate text-sm text-gray-700">{item.text}</p>
              <time className="shrink-0 text-xs text-gray-400">{item.time}</time>
            </li>
          ))}
        </ul>
      </section>
    </div>
  )
}
