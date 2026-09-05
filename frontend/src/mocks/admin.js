export const adminUsers = [
  { id: 'u1', name: 'Sandra Martínez', email: 'sandra.martinez@correo.ni', type: 'Comprador', region: 'Managua', since: 'Ene 2024', status: 'Activo', actions: ['Ver', 'Bloquear'] },
  { id: 'u2', name: 'Juan Ramón Martínez', email: 'cooperativa.cafenorte@correo.ni', type: 'Productor', region: 'Madriz', since: 'Mar 2023', status: 'Activo', actions: ['Ver', 'Bloquear'] },
  { id: 'u3', name: 'Doña María López', email: 'finca.elporvenir@correo.ni', type: 'Productor', region: 'Matagalpa', since: 'Ago 2022', status: 'Activo', actions: ['Ver', 'Bloquear'] },
  { id: 'u4', name: 'Roberto Sánchez', email: 'hacienda.sanjose@correo.ni', type: 'Productor', region: 'Chontales', since: 'Jun 2024', status: 'Pendiente', actions: ['Ver', 'Aprobar', 'Bloquear'] },
  { id: 'u5', name: 'Carlos Méndez', email: 'apiario.lassegovias@correo.ni', type: 'Productor', region: 'Nueva Segovia', since: 'Jul 2024', status: 'Pendiente', actions: ['Ver', 'Aprobar', 'Bloquear'] },
]

export const adminProducers = [
  { id: 'cooperativa-cafe-norte', name: 'Juan Ramón Martínez', farm: 'Cooperativa Café del Norte', location: 'San Juan de Río Coco, Madriz', active: true, productsCount: 15, rating: 4.9, since: 1998 },
  { id: 'frijoles-sebaco', name: 'Doña María López', farm: 'Finca El Porvenir', location: 'Sébaco, Matagalpa', active: true, productsCount: 8, rating: 4.8, since: 2005 },
  { id: 'ganaderia-chontales', name: 'Roberto Sánchez', farm: 'Hacienda San José', location: 'Juigalpa, Chontales', active: true, productsCount: 12, rating: 4.7, since: 1992 },
  { id: 'apiario-segovias', name: 'Carlos Méndez', farm: 'Apiario Las Segovias', location: 'Ocotal, Nueva Segovia', active: true, productsCount: 6, rating: 5.0, since: 2010 },
  { id: 'cacao-matiguas', name: 'Ana Isabel Rocha', farm: 'Finca La Esperanza', location: 'Matiguás, Matagalpa', active: true, productsCount: 7, rating: 4.9, since: 2015 },
  { id: 'plantaciones-rivas', name: 'Luis Alberto Torres', farm: 'Plantaciones del Sur', location: 'Rivas, Rivas', active: true, productsCount: 10, rating: 4.6, since: 2008 },
]

export const adminProducts = [
  { id: 'cafe-altura-shg', name: 'Café de Altura SHG', producer: 'Juan Ramón Martínez', category: 'Café', price: 180, status: 'Disponible', actions: ['Ver', 'Eliminar'] },
  { id: 'cafe-organico-jinotega', name: 'Café Orgánico Jinotega', producer: 'Pedro Antonio López', category: 'Café', price: 220, status: 'Disponible', actions: ['Ver', 'Eliminar'] },
  { id: 'frijol-rojo-seda', name: 'Frijol Rojo Seda', producer: 'Doña María López', category: 'Granos Básicos', price: 75, status: 'Disponible', actions: ['Ver', 'Eliminar'] },
  { id: 'frijol-negro', name: 'Frijol Negro', producer: 'Doña María López', category: 'Granos Básicos', price: 70, status: 'Disponible', actions: ['Ver', 'Eliminar'] },
  { id: 'platano-burro', name: 'Plátano Burro', producer: 'Luis Alberto Torres', category: 'Plátanos y Cocos', price: 12, status: 'Disponible', actions: ['Ver', 'Eliminar'] },
  { id: 'queso-seco-artesanal', name: 'Queso Seco Artesanal', producer: 'Roberto Sánchez', category: 'Lácteos y Quesos', price: 180, status: 'Disponible', actions: ['Ver', 'Eliminar'] },
  { id: 'cuajada-fresca', name: 'Cuajada Fresca', producer: 'Roberto Sánchez', category: 'Lácteos y Quesos', price: 85, status: 'Disponible', actions: ['Ver', 'Eliminar'] },
  { id: 'quesillo', name: 'Quesillo', producer: 'Familia Herrera', category: 'Lácteos y Quesos', price: 45, status: 'Disponible', actions: ['Ver', 'Eliminar'] },
  { id: 'carne-res-pastoreo', name: 'Carne de Res (Ganado de Pastoreo)', producer: 'Roberto Sánchez', category: 'Carnes y Ganadería', price: 280, status: 'Disponible', actions: ['Ver', 'Eliminar'] },
  { id: 'miel-cafe', name: 'Miel de Floración de Café', producer: 'Carlos Méndez', category: 'Miel y Apicultura', price: 160, status: 'Disponible', actions: ['Ver', 'Eliminar'] },
  { id: 'cacao-fino-aroma', name: 'Cacao Fino de Aroma', producer: 'Ana Isabel Rocha', category: 'Cacao y Chocolate', price: 280, status: 'Disponible', actions: ['Ver', 'Eliminar'] },
  { id: 'chocolate-70', name: 'Chocolate 70% Cacao', producer: 'Ana Isabel Rocha', category: 'Cacao y Chocolate', price: 120, status: 'Disponible', actions: ['Ver', 'Eliminar'] },
  { id: 'platano-burro', name: 'Plátano Burro', producer: 'Luis Alberto Torres', category: 'Plátanos y Cocos', price: 12, status: 'Disponible', actions: ['Ver', 'Eliminar'] },
  { id: 'queso-seco-artesanal', name: 'Queso Seco Artesanal', producer: 'Roberto Sánchez', category: 'Lácteos y Quesos', price: 180, status: 'Disponible', actions: ['Ver', 'Eliminar'] },
  { id: 'crema-leche', name: 'Crema de Leche', producer: 'Familia Herrera', category: 'Lácteos y Quesos', price: 60, status: 'Disponible', actions: ['Ver', 'Eliminar'] },
]

export const adminModerationReports = [
  {
    id: 'mod-1',
    title: 'Miel adulterada — Apiario Falso',
    severity: 'Alta',
    type: 'Producto',
    reportedBy: 'Sandra Martínez',
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
    detail: 'El teléfono y correo de la finca "Finca El Porvenir" no responden. Posible abandono de la cuenta.',
    status: 'pending',
  },
  {
    id: 'mod-3',
    title: 'Precio desactualizado por más de 60 días',
    severity: 'Baja',
    type: 'Producto',
    reportedBy: 'Sistema',
    time: 'Hace 1 día',
    detail: 'El producto "Maíz Blanco" no ha actualizado su precio en 62 días.',
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
    { icon: 'map', value: '153', label: 'Municipios activos', tone: 'brand' },
    { icon: 'chat_bubble', value: '12,480', label: 'Mensajes enviados', tone: 'blue' },
    { icon: 'person_add', value: '47', label: 'Nuevos hoy', tone: 'amber' },
    { icon: 'star', value: '4.9 ★', label: 'Valoración media', tone: 'red' },
  ],
}

export const recentRegistrations = [
  { name: 'Cooperativa Café El Paraíso', type: 'Productor', location: 'Jinotega', status: 'Pendiente', time: 'Hace 1 h' },
  { name: 'Carlos Mendoza', type: 'Comprador', location: 'Managua', status: 'Activo', time: 'Hace 3 h' },
  { name: 'Finca Los Robles', type: 'Productor', location: 'Matagalpa', status: 'Pendiente', time: 'Hace 6 h' },
  { name: 'Laura Estrada', type: 'Comprador', location: 'León', status: 'Activo', time: 'Hace 8 h' },
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
  { region: 'Matagalpa', count: 420 },
  { region: 'Jinotega', count: 310 },
  { region: 'Nueva Segovia', count: 280 },
  { region: 'Chontales', count: 250 },
  { region: 'Madriz', count: 190 },
]

export const categoryStats = [
  { category: 'Granos Básicos', count: 680, percent: 28 },
  { category: 'Café', count: 420, percent: 18 },
  { category: 'Frutas Tropicales', count: 380, percent: 15 },
  { category: 'Lácteos y Quesos', count: 210, percent: 9 },
  { category: 'Hortalizas', count: 180, percent: 7 },
  { category: 'Carnes y Ganadería', count: 150, percent: 6 },
  { category: 'Plátanos y Cocos', count: 120, percent: 5 },
  { category: 'Miel y Apicultura', count: 95, percent: 4 },
  { category: 'Cacao y Chocolate', count: 85, percent: 4 },
  { category: 'Procesados Artesanales', count: 75, percent: 3 },
]