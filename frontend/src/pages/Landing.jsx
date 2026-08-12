import Navbar from '../components/layout/Navbar'
import Footer from '../components/layout/Footer'

export default function Landing() {
  return (
    <div className="flex min-h-screen flex-col">
      <Navbar />
      <main id="inicio" className="flex-1" />
      <Footer />
    </div>
  )
}