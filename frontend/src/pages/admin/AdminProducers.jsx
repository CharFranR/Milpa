import ProducerCard from '../../components/product/ProducerCard'
import { adminProducers } from '../../mocks/admin'

export default function AdminProducers() {
  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold text-gray-900">Productores</h1>

      <div className="grid gap-5 sm:grid-cols-2 xl:grid-cols-3">
        {adminProducers.map((prod) => (
          <ProducerCard
            key={prod.id}
            producerId={prod.id}
            showActions
          />
        ))}
      </div>
    </div>
  )
}