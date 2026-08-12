import Navbar from '../components/layout/Navbar'
import Footer from '../components/layout/Footer'
import Hero from './landing/Hero'
import Stats from './landing/Stats'
import Categories from './landing/Categories'
import HowItWorks from './landing/HowItWorks'
import FeaturedProducts from './landing/FeaturedProducts'

export default function Landing() {
  return (
    <div className="flex min-h-screen flex-col">
      <Navbar />
      <main id="inicio" className="flex-1">
        <Hero />
        <Stats />
        <Categories />
        <HowItWorks />
        <FeaturedProducts />
      </main>
      <Footer />
    </div>
  )
}