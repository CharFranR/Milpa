import Navbar from '../components/layout/Navbar'
import Footer from '../components/layout/Footer'
import Hero from './landing/Hero'
import Stats from './landing/Stats'

export default function Landing() {
  return (
    <div className="flex min-h-screen flex-col">
      <Navbar />
      <main id="inicio" className="flex-1">
        <Hero />
        <Stats />
      </main>
      <Footer />
    </div>
  )
}