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
    city: 'Choachí',
    region: 'Cundinamarca',
    since: 2004,
    rating: 4.9,
    reviews: 238,
    productsCount: 12,
    specialties: ['Verduras orgánicas', 'Frutas de clima frío', 'Hierbas aromáticas', 'Sin pesticidas'],
  },
  {
    id: 'carlos-rios',
    name: 'Carlos Ríos',
    farm: 'Finca El Roble',
    city: 'Tenjo',
    region: 'Cundinamarca',
    since: 2011,
    rating: 4.8,
    reviews: 156,
    productsCount: 9,
    specialties: ['Frutas de clima frío', 'Granos', 'Aguacate hass'],
  },
  {
    id: 'ana-lucia-mora',
    name: 'Ana Lucía Mora',
    farm: 'Huerta La Violeta',
    city: 'Paipa',
    region: 'Boyacá',
    since: 2018,
    rating: 4.7,
    reviews: 89,
    productsCount: 6,
    specialties: ['Fresas', 'Frutos rojos', 'Cosecha semanal'],
  },
  {
    id: 'rosa-camargo',
    name: 'Rosa Camargo',
    farm: 'Apiario Doña Rosa',
    city: 'San Pedro',
    region: 'Nariño',
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
    name: 'Tomates Cherry Orgánicos',
    categoryId: 'verduras',
    producerId: 'maria-gonzalez',
    region: 'Cundinamarca',
    price: 12000,
    unit: 'kg',
    rating: 4.9,
    reviews: 34,
    available: true,
    tags: ['orgánico', 'sin pesticidas', 'cosecha semanal'],
    description:
      'Tomates cherry cultivados en invernadero con abonos orgánicos. Cosechados el mismo día de tu pedido para entregarlos en su punto óptimo de maduración. Dulces, firmes y perfectos para ensaladas o cocciones rápidas.',
  },
  {
    id: 'aguacate-hass',
    name: 'Aguacate Hass',
    categoryId: 'frutas',
    producerId: 'carlos-rios',
    region: 'Cundinamarca',
    price: 9000,
    unit: 'un',
    rating: 4.9,
    reviews: 41,
    available: true,
    tags: ['hass', 'maduración controlada', 'exportación'],
    description:
      'Aguacate hass de calidad exportación, con maduración controlada en finca. Cremoso, de cáscara gruesa y listo para consumir en 2 a 4 días. Ideal para guacamole y montaditos.',
  },
  {
    id: 'fresas-rubra',
    name: 'Fresas Rojas',
    categoryId: 'frutas',
    producerId: 'ana-lucia-mora',
    region: 'Boyacá',
    price: 15000,
    unit: 'bandeja',
    rating: 4.7,
    reviews: 19,
    available: true,
    tags: ['clima frío', 'cosecha semanal', 'sin agroquímicos'],
    description:
      'Fresas de la sabana de Paipa, cosechadas en la semana y seleccionadas una por una. Aromáticas y dulces, perfectas para postres, jugos o consumo directo.',
  },
  {
    id: 'miel-pura',
    name: 'Miel Pura de Abeja',
    categoryId: 'miel',
    producerId: 'rosa-camargo',
    region: 'Nariño',
    price: 28000,
    unit: '500 ml',
    rating: 5.0,
    reviews: 15,
    available: true,
    tags: ['100% pura', 'cruda', 'sin proceso térmico'],
    description:
      'Miel cruda de floración nativa de Nariño, extraída en frío y sin pasteurizar. Conserva todas sus enzimas y propiedades. Envasada en frasco de vidrio de 500 ml.',
  },
]

export const regions = [
  'Cundinamarca',
  'Boyacá',
  'Valle del Cauca',
  'Tolima',
  'Nariño',
  'Santander',
  'Antioquia',
  'Córdoba',
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

export function productsByCategory(categoryId) {
  return products.filter((p) => p.categoryId === categoryId)
}