export const buyerProfile = {
  name: 'Sandra Martínez',
  fullName: 'Sandra Martínez López',
  initials: 'SM',
  role: 'Comprador',
  city: 'Managua',
  region: 'Managua',
  email: 'sandra.martinez@correo.ni',
  phone: '+505 8888 4444',
}

export const homeStats = [
  { icon: 'favorite', value: '12', label: 'Favoritos', tone: 'red' },
  { icon: 'contact_phone', value: '5', label: 'Contactados', tone: 'blue' },
  { icon: 'search', value: '28', label: 'Búsquedas', tone: 'brand' },
  { icon: 'agriculture', value: '9', label: 'Productores vistos', tone: 'amber' },
]

export const quickActions = [
  { icon: 'search', label: 'Buscar productos', tab: 'marketplace' },
  { icon: 'location_on', label: 'Productores cercanos', tab: 'marketplace' },
  { icon: 'favorite', label: 'Mis favoritos', tab: 'favoritos' },
  { icon: 'chat_bubble', label: 'Mis mensajes', tab: 'mensajes' },
]

export const recommendedProductIds = ['cafe-altura-shg', 'frijol-rojo-seda', 'queso-seco-artesanal', 'miel-cafe']

export const favoriteProductIds = [
  'cafe-altura-shg',
  'frijol-rojo-seda',
  'platano-burro',
  'miel-pura',
  'lechuga-fresca',
  'queso-seco-artesanal',
]

export const recentActivity = [
  {
    icon: 'favorite',
    tone: 'red',
    text: 'Guardaste "Café de Altura SHG" en favoritos',
    time: 'Hace 2 horas',
  },
  {
    icon: 'chat_bubble',
    tone: 'blue',
    text: 'Contactaste a Juan Ramón Martínez de Cooperativa Café del Norte',
    time: 'Ayer',
  },
  {
    icon: 'search',
    tone: 'brand',
    text: 'Buscaste "frijol rojo seda Matagalpa"',
    time: 'Hace 2 días',
  },
]

export const conversations = [
  {
    id: 'cooperativa-cafe-norte',
    name: 'Juan Ramón Martínez',
    farm: 'Cooperativa Café del Norte',
    initials: 'JR',
    lastMessage: '¡Hola Sandra! Sí, tenemos café de altura SHG de la nueva cosecha disponible.',
    time: '10:23',
    unread: 2,
  },
  {
    id: 'frijoles-sebaco',
    name: 'Doña María López',
    farm: 'Finca El Porvenir',
    initials: 'ML',
    lastMessage: 'El frijol rojo seda está listo para envío',
    time: 'Ayer',
    unread: 0,
  },
  {
    id: 'ganaderia-chontales',
    name: 'Roberto Sánchez',
    farm: 'Hacienda San José',
    initials: 'RS',
    lastMessage: 'Tenemos queso seco y cuajada fresca esta semana',
    time: 'Lun',
    unread: 0,
  },
  {
    id: 'apiario-segovias',
    name: 'Carlos Méndez',
    farm: 'Apiario Las Segovias',
    initials: 'CM',
    lastMessage: 'La miel de floración de café ya está lista',
    time: 'Dom',
    unread: 0,
  },
]

export const activeChatMessages = [
  { from: 'me', text: '¡Hola Juan! Me interesa el Café de Altura SHG para mi cafetería. ¿Tienen disponibilidad de 5 kg para esta semana?' },
  { from: 'them', text: '¡Hola Sandra! Sí, tenemos café de altura SHG de la nueva cosecha disponible. Son C$180/kg. ¿Cuántos kilos necesitas?' },
  { from: 'me', text: '¡Perfecto! Me interesan unos 5 kg para mi cafetería. ¿Podrían hacer el envío a Managua esta semana?' },
  { from: 'them', text: '¡Claro que sí! Lo preparamos y se lo enviamos el martes. Quedamos en contacto para coordinar la entrega.' },
]