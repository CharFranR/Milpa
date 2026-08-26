import ProductCard from '../../components/product/ProductCard'
import { buyerProfile, favoriteProductIds } from '../../mocks/buyer'

export default function BuyerFavorites() {
  return (
    <div className="space-y-6">
      <header>
        <h1 className="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">
          Mis favoritos
        </h1>
        <p className="mt-1 text-sm text-gray-500">
          Hola {buyerProfile.name}, tienes {favoriteProductIds.length} productos guardados.
        </p>
      </header>

      <section aria-label="Productos favoritos">
        <div className="grid gap-5 sm:grid-cols-2 xl:grid-cols-3">
          {favoriteProductIds.map((productId) => (
            <ProductCard key={productId} productId={productId} initiallyFavorite />
          ))}
        </div>
      </section>
    </div>
  )
}
