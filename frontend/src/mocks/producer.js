export const producerProfile = {
  id: 'cooperativa-cafe-norte',
  name: 'Juan Ramón Martínez',
  farm: 'Cooperativa Café del Norte',
  city: 'San Juan de Río Coco',
  region: 'Madriz',
  since: 1998,
  rating: 4.9,
  reviews: 312,
  email: 'cooperativa.cafenorte@correo.ni',
  phone: '+505 8777 3333',
  schedule: 'Lun–Vie: 7am – 4pm',
  specialties: ['Café de altura SHG', 'Café orgánico', 'Variedades Caturra y Catuai', 'Proceso lavado y honey'],
  social: {
    instagram: '@cooperativacafenorte',
    facebook: 'Cooperativa Café del Norte',
  },
}

export const producerStats = [
  { icon: 'inventory_2', value: '15', label: 'Productos activos', trend: '+3 este mes', tone: 'brand' },
  { icon: 'inbox', value: '52', label: 'Solicitudes recibidas', trend: '+8 esta semana', tone: 'blue' },
  { icon: 'contact_phone', value: '35', label: 'Compradores contactados', trend: 'Total', tone: 'amber' },
  { icon: 'star', value: '4.9 ★', label: 'Valoración promedio', trend: '312 reseñas', tone: 'red' },
]

export const producerProducts = [
  { id: 'cafe-altura-shg', name: 'Café de Altura SHG', category: 'Café', price: 180, unit: 'kg', available: true },
  { id: 'cafe-organico-jinotega', name: 'Café Orgánico Jinotega', category: 'Café', price: 220, unit: 'kg', available: true },
  { id: 'miel-cafe', name: 'Miel de Floración de Café', category: 'Miel y Apicultura', price: 160, unit: '500 ml', available: true },
  { id: 'polen-apicola', name: 'Polen Apícola', category: 'Miel y Apicultura', price: 180, unit: '250 g', available: true },
]

export const producerRequests = [
  {
    id: 'req-1',
    productId: 'cafe-altura-shg',
    buyer: { name: 'Cafetería El Grano', avatar: 'CG', tone: 'red' },
    product: 'Café de Altura SHG',
    qty: '20 kg',
    status: 'pending',
    time: 'Hace 30 min',
    message: 'Necesitamos café para el menú de la próxima semana. ¿Disponibilidad inmediata?',
  },
  {
    id: 'req-2',
    productId: 'miel-cafe',
    buyer: { name: 'Sandra Martínez', avatar: 'SM', tone: 'brand' },
    product: 'Miel de Floración de Café',
    qty: '5 frascos (500 ml)',
    status: 'pending',
    time: 'Hace 2 h',
    message: 'Hola, vi su miel de floración de café y me gustaría comprarla para mi cafetería. ¿Está disponible?',
  },
  {
    id: 'req-3',
    productId: 'cafe-altura-shg',
    buyer: { name: 'Tostaduría El Grano de Oro', avatar: 'TGO', tone: 'amber' },
    product: 'Café de Altura SHG',
    qty: '50 kg',
    status: 'responded',
    time: 'Ayer',
    message: 'Necesitamos para tostado de la próxima semana. ¿Podemos coordinar entrega el martes?',
  },
  {
    id: 'req-4',
    productId: 'cafe-altura-shg',
    buyer: { name: 'Exportadora Café Nicaragua', avatar: 'ECN', tone: 'blue' },
    product: 'Café de Altura SHG',
    qty: '100 kg',
    status: 'responded',
    time: 'Hace 3 días',
    message: 'Perfecto, gracias por la rapidez en la cotización.',
  },
]

export const producerConversations = [
  {
    id: 'conv-1',
    buyer: { name: 'Cafetería El Grano', avatar: 'CG', farm: 'Cafetería', tone: 'red' },
    lastMessage: '¿Cuándo puedo pasar a recoger el pedido?',
    time: '10:23',
    unread: 2,
  },
  {
    id: 'conv-2',
    buyer: { name: 'Sandra Martínez', avatar: 'SM', farm: 'Compradora', tone: 'brand' },
    lastMessage: '¡Perfecto! Me interesan unos 5 kg para mi cafetería. ¿Podrían hacer el envío a Managua esta semana?',
    time: '10:25',
    unread: 1,
  },
  {
    id: 'conv-3',
    buyer: { name: 'Tostaduría El Grano de Oro', avatar: 'TGO', farm: 'Tostaduría', tone: 'amber' },
    lastMessage: 'El café llegó en perfectas condiciones, gracias',
    time: 'Lun',
    unread: 0,
  },
]

export const producerBusiness = {
  farmPhoto: '',
  farmName: 'Cooperativa Café del Norte',
  location: 'San Juan de Río Coco, Madriz',
  since: 1998,
  contact: {
    email: 'cooperativa.cafenorte@correo.ni',
    phone: '+505 8777 3333',
    schedule: 'Lun–Vie: 7am – 4pm',
  },
  specialties: ['Café de altura SHG', 'Café orgánico', 'Variedades Caturra y Catuai', 'Proceso lavado y honey'],
  social: {
    instagram: '@cooperativacafenorte',
    facebook: 'Cooperativa Café del Norte',
  },
}