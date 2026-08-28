export const buyerProfile = {
  name: 'Sandra Campos',
  fullName: 'Sandra Campos Ruiz',
  initials: 'SC',
  role: 'Comprador',
  city: 'Medellín',
  region: 'Antioquia',
  email: 'sandra@correo.com',
  phone: '+57 310 555 0199',
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

export const recommendedProductIds = ['tomates-cherry', 'aguacate-hass', 'fresas-rubra', 'miel-pura']

export const favoriteProductIds = [
  'tomates-cherry',
  'aguacate-hass',
  'fresas-rubra',
  'miel-pura',
  'lechuga-butter',
  'queso-campesino',
]

export const recentActivity = [
  {
    icon: 'favorite',
    tone: 'red',
    text: 'Guardaste "Tomates Cherry Orgánicos" en favoritos',
    time: 'Hace 2 horas',
  },
  {
    icon: 'chat_bubble',
    tone: 'blue',
    text: 'Contactaste a María González de Finca La Esperanza',
    time: 'Ayer',
  },
  {
    icon: 'search',
    tone: 'brand',
    text: 'Buscaste "aguacate hass boyacá"',
    time: 'Hace 2 días',
  },
]

export const conversations = [
  {
    id: 'maria-gonzalez',
    name: 'María González',
    farm: 'Finca La Esperanza',
    initials: 'MG',
    lastMessage: 'Perfecto, nos vemos el martes en la plaza',
    time: '10:23',
    unread: 2,
  },
  {
    id: 'carlos-rios',
    name: 'Carlos Ríos',
    farm: 'Finca El Roble',
    initials: 'CR',
    lastMessage: 'El aguacate ya está en su punto, te lo aparto',
    time: 'Ayer',
    unread: 0,
  },
  {
    id: 'ana-lucia-mora',
    name: 'Ana Lucía Mora',
    farm: 'Huerta La Violeta',
    initials: 'AM',
    lastMessage: 'Esta semana entra cosecha nueva de fresas',
    time: 'Lun',
    unread: 0,
  },
  {
    id: 'jose-ramirez',
    name: 'José Ramírez',
    farm: 'Huerto El Progreso',
    initials: 'JR',
    lastMessage: 'Gracias por tu compra, ¡buen provecho!',
    time: 'Dom',
    unread: 0,
  },
]

export const activeChatMessages = [
  { from: 'them', text: '¡Hola Sandra! Sí, tenemos tomates cherry disponibles esta semana.' },
  { from: 'me', text: '¡Hola María! Me interesan unos 3 kg para el restaurante.' },
  { from: 'me', text: '¿Podría ser la entrega el martes en la mañana?' },
]
