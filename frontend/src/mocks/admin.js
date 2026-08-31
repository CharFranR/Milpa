export const adminUsers = [
  { id: 'u1', name: 'Sandra Campos', email: 'sandra@correo.com', type: 'Comprador', region: 'Antioquia', since: 'Ene 2024', status: 'Activo', actions: ['Ver', 'Bloquear'] },
  { id: 'u2', name: 'María González', email: 'finca@laesperanza.co', type: 'Productor', region: 'Cundinamarca', since: 'Mar 2023', status: 'Activo', actions: ['Ver', 'Bloquear'] },
  { id: 'u3', name: 'Carlos Ríos', email: 'carlos@elroble.co', type: 'Productor', region: 'Cundinamarca', since: 'Ago 2022', status: 'Activo', actions: ['Ver', 'Bloquear'] },
  { id: 'u4', name: 'Juliana Morales', email: 'juliana@correo.com', type: 'Comprador', region: 'Valle del Cauca', since: 'Jun 2024', status: 'Pendiente', actions: ['Ver', 'Aprobar', 'Bloquear'] },
  { id: 'u5', name: 'Coop Agro Boyacá', email: 'contacto@agroboyaca.co', type: 'Productor', region: 'Boyacá', since: 'Jul 2024', status: 'Pendiente', actions: ['Ver', 'Aprobar', 'Bloquear'] },
]

export const adminProducers = [
  { id: 'maria-gonzalez', name: 'María González', farm: 'Finca La Esperanza', location: 'Choachí, Cundinamarca', active: true, productsCount: 12, rating: 4.9, since: 2004 },
  { id: 'carlos-rios', name: 'Carlos Ríos', farm: 'Finca El Roble', location: 'Tenjo, Cundinamarca', active: true, productsCount: 9, rating: 4.8, since: 2011 },
  { id: 'ana-lucia-mora', name: 'Ana Lucía Mora', farm: 'Huerta La Violeta', location: 'Paipa, Boyacá', active: true, productsCount: 6, rating: 4.7, since: 2018 },
  { id: 'rosa-camargo', name: 'Rosa Camargo', farm: 'Apiario Doña Rosa', location: 'San Pedro, Nariño', active: true, productsCount: 5, rating: 5.0, since: 2010 },
  { id: 'coop-boyaca', name: 'Cooperativa Agro Boyacá', farm: 'Cooperativa Agro Boyacá', location: 'Tunja, Boyacá', active: false, productsCount: 8, rating: 4.5, since: 2023 },
  { id: 'finca-robles', name: 'Finca Los Robles', farm: 'Finca Los Robles', location: 'Manizales, Caldas', active: false, productsCount: 4, rating: 4.2, since: 2022 },
]

export const adminProducts = [
  { id: 'tomates-cherry', name: 'Tomates Cherry Orgánicos', producer: 'María González', category: 'Verduras', price: 12000, status: 'Disponible', actions: ['Ver', 'Eliminar'] },
  { id: 'aguacate-hass', name: 'Aguacate Hass', producer: 'Carlos Ríos', category: 'Frutas', price: 9000, status: 'Disponible', actions: ['Ver', 'Eliminar'] },
  { id: 'fresas-rubra', name: 'Fresas Rojas', producer: 'Ana Lucía Mora', category: 'Frutas', price: 15000, status: 'Disponible', actions: ['Ver', 'Eliminar'] },
  { id: 'miel-pura', name: 'Miel Pura de Abeja', producer: 'Rosa Camargo', category: 'Miel', price: 28000, status: 'Disponible', actions: ['Ver', 'Eliminar'] },
  { id: 'lechuga-butter', name: 'Lechuga Butter Hidropónica', producer: 'María González', category: 'Verduras', price: 3500, status: 'Disponible', actions: ['Ver', 'Eliminar'] },
  { id: 'queso-campesino', name: 'Queso Campesino Artesanal', producer: 'María González', category: 'Lácteos', price: 16000, status: 'Disponible', actions: ['Ver', 'Eliminar'] },
  { id: 'zanahorias-baby', name: 'Zanahorias Baby Orgánicas', producer: 'María González', category: 'Verduras', price: 8500, status: 'Agotado', actions: ['Ver', 'Eliminar'] },
  { id: 'cilantro-fresco', name: 'Cilantro Fresco', producer: 'María González', category: 'Hierbas', price: 2000, status: 'Disponible', actions: ['Ver', 'Eliminar'] },
]

export const adminModerationReports = [
  {
    id: 'mod-1',
    title: 'Miel adulterada — Apiario Falso',
    severity: 'Alta',
    type: 'Producto',
    reportedBy: 'Sandra Campos',
    time: 'Hace 2 h',
    detail: 'El producto anunciado como "Miel Pura de Abeja" contiene jarabe de maíz según análisis independiente.',
    status: 'pending',
  },
  {
    id: 'mod-2',
    title: 'Información de contacto incorrecta',
    severity: 'Media',
    type: 'Productor',
    reportedBy: 'Carlos Ríos',
    time: 'Hace 5 h',
    detail: 'El teléfono y correo de la finca "Huerta La Violeta" no responden. Posible abandono de la cuenta.',
    status: 'pending',
  },
  {
    id: 'mod-3',
    title: 'Precio desactualizado por más de 60 días',
    severity: 'Baja',
    type: 'Producto',
    reportedBy: 'Sistema',
    time: 'Hace 1 día',
    detail: 'El producto "Maíz Amarillo Tradicional" no ha actualizado su precio en 62 días.',
    status: 'resolved',
  },
]

export const adminStats = {
  primary: [
    { icon: 'group', value: '5,842', label: 'Usuarios totales', sub: '↑ +124 esta semana', tone: 'blue' },
    { icon: 'agriculture', value: '2,401', label: 'Productores activos', sub: '↑ +18 este mes', tone: 'brand' },
    { icon: 'inventory_2', value: '18,534', label: 'Productos publicados', sub: '↑ +340 este mes', tone: 'amber' },
    { icon: 'flag', value: '2', label: 'Reportes pendientes', sub: '↓ 3 resueltos hoy', tone: 'red' },
  ],
  secondary: [
    { icon: 'map', value: '340', label: 'Municipios activos', tone: 'brand' },
    { icon: 'chat_bubble', value: '12,480', label: 'Mensajes enviados', tone: 'blue' },
    { icon: 'person_add', value: '47', label: 'Nuevos hoy', tone: 'amber' },
    { icon: 'star', value: '4.9 ★', label: 'Valoración media', tone: 'red' },
  ],
}

export const recentRegistrations = [
  { name: 'Cooperativa Agro Boyacá', type: 'Productor', location: 'Tunja', status: 'Pendiente', time: 'Hace 1 h' },
  { name: 'Carlos Mendoza', type: 'Comprador', location: 'Bogotá', status: 'Activo', time: 'Hace 3 h' },
  { name: 'Finca Los Robles', type: 'Productor', location: 'Manizales', status: 'Pendiente', time: 'Hace 6 h' },
  { name: 'Laura Estrada', type: 'Comprador', location: 'Medellín', status: 'Activo', time: 'Hace 8 h' },
]

export const pendingReports = [
  { title: 'Miel adulterada — Apiario Falso', severity: 'Alta', color: 'red' },
  { title: 'Información de contacto incorrecta', severity: 'Media', color: 'yellow' },
  { title: 'Precio desactualizado por más de 60 días', severity: 'Baja', color: 'gray' },
]

export const growthStats = [
  { label: 'Nuevos usuarios', value: '+482', total: '5,842 total', progress: 482 / 5842 },
  { label: 'Nuevos productores', value: '+38', total: '2,401 total', progress: 38 / 2401 },
  { label: 'Nuevos productos', value: '+890', total: '18,534 total', progress: 890 / 18534 },
]

export const regionRanking = [
  { region: 'Cundinamarca', count: 420 },
  { region: 'Boyacá', count: 310 },
  { region: 'Valle del Cauca', count: 280 },
  { region: 'Antioquia', count: 250 },
  { region: 'Nariño', count: 190 },
]

export const categoryStats = [
  { category: 'Verduras', count: 520, percent: 28 },
  { category: 'Frutas', count: 340, percent: 18 },
  { category: 'Lácteos', count: 210, percent: 11 },
  { category: 'Granos', count: 180, percent: 10 },
  { category: 'Hierbas', count: 140, percent: 8 },
  { category: 'Carnes', count: 95, percent: 5 },
  { category: 'Miel', count: 72, percent: 4 },
  { category: 'Flores', count: 58, percent: 3 },
]