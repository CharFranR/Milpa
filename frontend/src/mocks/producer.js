export const producerProfile = {
  id: 'maria-gonzalez',
  name: 'María González',
  farm: 'Finca La Esperanza',
  city: 'Choachí',
  region: 'Cundinamarca',
  since: 2004,
  rating: 4.9,
  reviews: 238,
  email: 'finca@laesperanza.co',
  phone: '+57 310 555 0101',
  schedule: 'Lun–Vie: 7am – 5pm',
  specialties: ['Verduras orgánicas', 'Frutas de clima frío', 'Hierbas aromáticas', 'Sin pesticidas', 'Cosecha semanal'],
  social: {
    instagram: '@fincalaesperanza',
    facebook: 'Finca La Esperanza',
  },
}

export const producerStats = [
  { icon: 'inventory_2', value: '12', label: 'Productos activos', trend: '+2 este mes', tone: 'brand' },
  { icon: 'inbox', value: '48', label: 'Solicitudes recibidas', trend: '+12 esta semana', tone: 'blue' },
  { icon: 'contact_phone', value: '24', label: 'Compradores contactados', trend: 'Total', tone: 'amber' },
  { icon: 'star', value: '4.9 ★', label: 'Valoración promedio', trend: '238 reseñas', tone: 'red' },
]

export const producerProducts = [
  { id: 'tomates-cherry', name: 'Tomates Cherry Orgánicos', category: 'Verduras', price: 12000, unit: 'kg', available: true },
  { id: 'lechuga-butter', name: 'Lechuga Butter Hidropónica', category: 'Verduras', price: 3500, unit: 'un', available: true },
  { id: 'queso-campesino', name: 'Queso Campesino Artesanal', category: 'Lácteos', price: 16000, unit: 'kg', available: true },
  { id: 'cilantro-fresco', name: 'Cilantro Fresco', category: 'Hierbas', price: 2000, unit: 'atado', available: true },
  { id: 'zanahorias-baby', name: 'Zanahorias Baby Orgánicas', category: 'Verduras', price: 8500, unit: 'kg', available: false },
]

export const producerRequests = [
  {
    id: 'req-1',
    buyer: { name: 'Restaurante La Cosecha', avatar: 'LC', tone: 'red' },
    product: 'Tomates Cherry Orgánicos',
    qty: '10 kg',
    status: 'pending',
    time: 'Hace 30 min',
    message: 'Me interesan para el menú de la próxima semana. ¿Disponibilidad inmediata?',
  },
  {
    id: 'req-2',
    buyer: { name: 'Sandra Campos', avatar: 'SC', tone: 'brand' },
    product: 'Zanahorias Baby Orgánicas',
    qty: '3 kg',
    status: 'pending',
    time: 'Hace 2 h',
    message: 'Hola, vi sus zanahorias y me gustaría comprarlas para mi familia. ¿Están frescas?',
  },
  {
    id: 'req-3',
    buyer: { name: 'Tienda Natural Verde', avatar: 'TN', tone: 'amber' },
    product: 'Tomates Cherry Orgánicos',
    qty: '25 kg',
    status: 'responded',
    time: 'Ayer',
    message: 'Necesitamos para reabastecimiento. ¿Podemos coordinar entrega el martes?',
  },
  {
    id: 'req-4',
    buyer: { name: 'Juliana Morales', avatar: 'JM', tone: 'blue' },
    product: 'Tomates Cherry Orgánicos',
    qty: '2 kg',
    status: 'responded',
    time: 'Hace 3 días',
    message: 'Perfecto, gracias por la rapidez.',
  },
]

export const producerConversations = [
  {
    id: 'conv-1',
    buyer: { name: 'Restaurante La Cosecha', avatar: 'LC', farm: 'Restaurante', tone: 'red' },
    lastMessage: '¿Cuándo puedo pasar a recoger?',
    time: '10:23',
    unread: 2,
  },
  {
    id: 'conv-2',
    buyer: { name: 'Sandra Campos', avatar: 'SC', farm: 'Compradora', tone: 'brand' },
    lastMessage: 'Perfecto, nos vemos el martes',
    time: 'Ayer',
    unread: 0,
  },
  {
    id: 'conv-3',
    buyer: { name: 'Tienda Natural Verde', avatar: 'TN', farm: 'Tienda orgánica', tone: 'amber' },
    lastMessage: 'Los tomates llegaron perfectos, gracias',
    time: 'Lun',
    unread: 0,
  },
]

export const producerBusiness = {
  farmPhoto: '',
  farmName: 'Finca La Esperanza',
  location: 'Choachí, Cundinamarca',
  since: 2004,
  contact: {
    email: 'finca@laesperanza.co',
    phone: '+57 310 555 0101',
    schedule: 'Lun–Vie: 7am – 5pm',
  },
  specialties: ['Verduras orgánicas', 'Frutas de clima frío', 'Hierbas aromáticas', 'Sin pesticidas', 'Cosecha semanal'],
  social: {
    instagram: '@fincalaesperanza',
    facebook: 'Finca La Esperanza',
  },
}