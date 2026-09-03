export const categories = [
  { id: 'frutas', name: 'Frutas', icon: 'eco', count: 340 },
  { id: 'verduras', name: 'Verduras', icon: 'grass', count: 520 },
  { id: 'granos', name: 'Granos', icon: 'bakery_dining', count: 180 },
  { id: 'lacteos', name: 'Lácteos y huevos', icon: 'egg', count: 210 },
  { id: 'carnes', name: 'Carnes', icon: 'kebab_dining', count: 95 },
  { id: 'hierbas', name: 'Hierbas', icon: 'spa', count: 140 },
  { id: 'miel', name: 'Miel y derivados', icon: 'water_drop', count: 72 },
  { id: 'flores', name: 'Flores', icon: 'local_florist', count: 58 },
]

export const producers = [
  {
    id: 'maria-gonzalez',
    name: 'María González',
    farm: 'Finca La Esperanza',
    city: 'Matagalpa',
    region: 'Matagalpa',
    since: 2004,
    rating: 4.9,
    reviews: 238,
    productsCount: 12,
    specialties: ['Verduras frescas', 'Lácteos artesanales', 'Hierbas de olor', 'Sin pesticidas'],
  },
  {
    id: 'carlos-rios',
    name: 'Carlos Ríos',
    farm: 'Hacienda El Roble',
    city: 'Nagarote',
    region: 'León',
    since: 2011,
    rating: 4.8,
    reviews: 156,
    productsCount: 9,
    specialties: ['Frutas tropicales', 'Granos básicos', 'Ganadería de pastoreo'],
  },
  {
    id: 'ana-lucia-mora',
    name: 'Ana Lucía Mora',
    farm: 'Huerta La Violeta',
    city: 'San Rafael del Norte',
    region: 'Jinotega',
    since: 2018,
    rating: 4.7,
    reviews: 89,
    productsCount: 6,
    specialties: ['Fresas', 'Flores de altura', 'Aves de corral'],
  },
  {
    id: 'rosa-camargo',
    name: 'Rosa Camargo',
    farm: 'Apiario Doña Rosa',
    city: 'Condega',
    region: 'Estelí',
    since: 2010,
    rating: 5.0,
    reviews: 74,
    productsCount: 5,
    specialties: ['Miel pura', 'Apicultura', 'Polen'],
  },
]

export const products = [
  {
    id: 'tomates-cherry',
    name: 'Tomates de Mesa Orgánicos',
    categoryId: 'verduras',
    producerId: 'maria-gonzalez',
    region: 'Matagalpa',
    price: 80,
    unit: 'kg',
    rating: 4.9,
    reviews: 34,
    available: true,
    tags: ['orgánico', 'sin pesticidas', 'cosecha semanal'],
    description:
      'Tomates de mesa cultivados con abonos orgánicos en el clima fresco de Matagalpa. Cosechados el mismo día de tu pedido para entregarlos en su punto óptimo. Firmes y perfectos para ensaladas o guisos.',
  },
  {
    id: 'aguacate-mantequilla',
    name: 'Aguacate Mantequilla',
    categoryId: 'frutas',
    producerId: 'carlos-rios',
    region: 'León',
    price: 35,
    unit: 'un',
    rating: 4.9,
    reviews: 41,
    available: true,
    tags: ['criollo', 'maduración natural', 'temporada'],
    description:
      'Aguacate mantequilla tradicional nicaragüense, cremoso y de semilla pequeña. Listo para consumir en 2 a 4 días. Ideal para acompañar tus comidas típicas o hacer guacamole.',
  },
  {
    id: 'fresas-rubra',
    name: 'Fresas de Altura',
    categoryId: 'frutas',
    producerId: 'ana-lucia-mora',
    region: 'Jinotega',
    price: 85,
    unit: 'bandeja',
    rating: 4.7,
    reviews: 19,
    available: true,
    tags: ['clima frío', 'cosecha semanal', 'jinotega'],
    description:
      'Fresas cultivadas en las montañas de Jinotega, seleccionadas una por una. Dulces y jugosas, perfectas para postres, licuados o consumo directo.',
  },
  {
    id: 'miel-pura',
    name: 'Miel Pura de Abeja',
    categoryId: 'miel',
    producerId: 'rosa-camargo',
    region: 'Estelí',
    price: 180,
    unit: '500 ml',
    rating: 5.0,
    reviews: 15,
    available: true,
    tags: ['100% pura', 'cruda', 'norteña'],
    description:
      'Miel cruda de floración nativa de Las Segovias, extraída sin procesos térmicos. Conserva todas sus propiedades y enzimas. Envasada en frasco de vidrio.',
  },
  {
    id: 'lechuga-butter',
    name: 'Lechuga Fresca',
    categoryId: 'verduras',
    producerId: 'maria-gonzalez',
    region: 'Matagalpa',
    price: 25,
    unit: 'un',
    rating: 4.8,
    reviews: 22,
    available: true,
    tags: ['fresca', 'cosecha del día'],
    description:
      'Lechuga cultivada con agua limpia de manantial. Hojas crujientes y tiernas, ideales para ensaladas. Se cosecha la mañana misma del despacho.',
  },
  {
    id: 'queso-seco',
    name: 'Queso Seco Artesanal',
    categoryId: 'lacteos',
    producerId: 'maria-gonzalez',
    region: 'Matagalpa',
    price: 150,
    unit: 'kg',
    rating: 4.9,
    reviews: 31,
    available: true,
    tags: ['artesanal', 'leche pura', 'salado'],
    description:
      'Auténtico queso seco nicaragüense, elaborado con leche pura de vaca. Punto de sal exacto, consistencia firme, ideal para rallar sobre maduro frito o acompañar gallopinto.',
  },
  {
    id: 'cilantro-fresco',
    name: 'Culantro y Cilantro Fresco',
    categoryId: 'hierbas',
    producerId: 'maria-gonzalez',
    region: 'Matagalpa',
    price: 15,
    unit: 'atado',
    rating: 4.7,
    reviews: 18,
    available: true,
    tags: ['recién cortado', 'huerto casero'],
    description:
      'Atado mixto de cilantro y culantro de patio, con aroma intenso indispensable para tus sopas y picadillos. Cultivado sin químicos.',
  },
  {
    id: 'frijol-rojo',
    name: 'Frijol Rojo Seda',
    categoryId: 'granos',
    producerId: 'carlos-rios',
    region: 'León',
    price: 75,
    unit: 'kg',
    rating: 4.8,
    reviews: 112,
    available: true,
    tags: ['grano nuevo', 'cosecha reciente', 'seda'],
    description:
      'Frijol rojo seda de cosecha nueva. Grano limpio, que ablanda rápido al fuego, con el sabor tradicional perfecto para una buena sopa de frijoles o gallopinto.',
  },
  {
    id: 'carne-de-res',
    name: 'Posta de Res (Ganado de Pastoreo)',
    categoryId: 'carnes',
    producerId: 'carlos-rios',
    region: 'León',
    price: 320,
    unit: 'kg',
    rating: 4.3,
    reviews: 11,
    available: true,
    tags: ['ganado libre', 'corte magro'],
    description:
      'Posta de res de ganado criado en pastoreo natural en el occidente del país. Corte magro y fresco, despachado el mismo día. Ideal para desmenuzar o tapar.',
  },
  {
    id: 'huevos-campo',
    name: 'Huevos de Amor (Gallina India)',
    categoryId: 'lacteos',
    producerId: 'ana-lucia-mora',
    region: 'Jinotega',
    price: 190,
    unit: 'cajilla',
    rating: 4.8,
    reviews: 27,
    available: true,
    tags: ['gallinas de patio', 'yema amarilla'],
    description:
      'Huevos de gallinas indias de patio, alimentadas con maíz y lombrices. Yema de color intenso y excelente sabor. Empacados en cajilla de 30 unidades.',
  },
  {
    id: 'rosas-rojas',
    name: 'Rosas de Clima Fresco',
    categoryId: 'flores',
    producerId: 'ana-lucia-mora',
    region: 'Jinotega',
    price: 220,
    unit: 'docena',
    rating: 4.9,
    reviews: 9,
    available: true,
    tags: ['cultivo de altura', 'tallo largo'],
    description:
      'Rosas rojas cultivadas en el clima fresco de Jinotega, con tallos largos y botón cerrado. Cortadas el día del despacho para garantizar su duración en florero.',
  },
  {
    id: 'polen-apicola',
    name: 'Polen Apícola Norteño',
    categoryId: 'miel',
    producerId: 'rosa-camargo',
    region: 'Estelí',
    price: 160,
    unit: '250 g',
    rating: 5.0,
    reviews: 8,
    available: true,
    tags: ['rico en vitaminas', 'floración silvestre'],
    description:
      'Polen recolectado por abejas de la rica floración silvestre esteliana. Un superalimento lleno de energía. Deshidratado suavemente para conservar sus nutrientes.',
  },
]

export const regions = [
  'León',
  'Managua',
  'Jinotega',
  'Chinandega',
  'Matagalpa',
  'Nueva Segovia',
  'Estelí',
  'Rivas',
  'Otra región',
]

export function producerById(id) {
  return producers.find((p) => p.id === id)
}

export function categoryById(id) {
  return categories.find((c) => c.id === id)
}

export function productById(id) {
  return products.find((p) => p.id === id)
}

export function productImageUrl(productId) {
  return `/images/products/${productId}.jpg`
}

export function productsByCategory(categoryId) {
  return products.filter((p) => p.categoryId === categoryId)
}