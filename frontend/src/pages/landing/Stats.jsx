import StatCard from '../../components/StatCard'
import { landingStats } from '../../mocks/content'

export default function Stats() {
  return (
    <section className="bg-white">
      <div className="mx-auto grid max-w-7xl grid-cols-1 gap-4 px-4 py-14 sm:grid-cols-2 sm:px-6 lg:grid-cols-4 lg:px-8">
        {landingStats.map((stat, i) => (
          <StatCard key={stat.label} {...stat} index={i} />
        ))}
      </div>
    </section>
  )
}