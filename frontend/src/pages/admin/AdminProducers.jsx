import ProducerCard from '../../components/product/ProducerCard'
import { adminProducers } from '../../mocks/admin'

export default function AdminProducers() {
  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold text-gray-900">Productores</h1>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 max-w-full">
        {adminProducers.map((prod) => (
          <ProducerCard
            key={prod.id}
            producer={prod}
            showActions
          />
        ))}
      </div>
    </div>
  )
}