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
  {
    id: 'lechuga-butter',
    name: 'Lechuga Butter Hidropónica',
    categoryId: 'verduras',
    producerId: 'maria-gonzalez',
    region: 'Cundinamarca',
    price: 3500,
    unit: 'un',
    rating: 4.8,
    reviews: 22,
    available: true,
    tags: ['hidropónica', 'cosecha del día'],
    description:
      'Lechuga butter cultivada en sistema hidropónico con agua filtrada. Hojas tiernas y de sabor suave, ideales para ensaladas frescas. Se cosecha la mañana misma del despacho.',
  },
  {
    id: 'queso-campesino',
    name: 'Queso Campesino Artesanal',
    categoryId: 'lacteos',
    producerId: 'maria-gonzalez',
    region: 'Cundinamarca',
    price: 16000,
    unit: 'kg',
    rating: 4.9,
    reviews: 31,
    available: true,
    tags: ['artesanal', 'leche fresca'],
    description:
      'Queso campesino elaborado artesanalmente con leche fresca de fincas vecinas. Textura cremosa y sabor suave, perfecto para arepas, asados o mesa de desayuno.',
  },
  {
    id: 'cilantro-fresco',
    name: 'Cilantro Fresco',
    categoryId: 'hierbas',
    producerId: 'maria-gonzalez',
    region: 'Cundinamarca',
    price: 2000,
    unit: 'atado',
    rating: 4.7,
    reviews: 18,
    available: true,
    tags: ['recién cortado', 'sin agroquímicos'],
    description:
      'Cilantro recién cortado de huerta, con aroma intenso y tallos firmes. Cultivado sin agroquímicos y amarrado a mano en atados generosos.',
  },
  {
    id: 'maiz-amarillo',
    name: 'Maíz Amarillo Tradicional',
    categoryId: 'granos',
    producerId: 'carlos-rios',
    region: 'Cundinamarca',
    price: 4200,
    unit: 'kg',
    rating: 4.2,
    reviews: 12,
    available: true,
    tags: ['grano seco', 'variedad criolla'],
    description:
      'Maíz amarillo secado al sol, de la variedad tradicional usada para arepas, mazamorra y mote. Molido a pedido si lo prefieres en harina.',
  },
  {
    id: 'carne-de-res',
    name: 'Carne de Res a Pasto',
    categoryId: 'carnes',
    producerId: 'carlos-rios',
    region: 'Cundinamarca',
    price: 32000,
    unit: 'kg',
    rating: 4.3,
    reviews: 11,
    available: true,
    tags: ['ganado de pastoreo', 'despacho del día'],
    description:
      'Carne de res de ganado criado a pasto, madurada y despachada fresca el mismo día. Cortes disponibles según temporada: punta de anca, muchacho y sobrebarriga.',
  },
  {
    id: 'huevos-campo',
    name: 'Huevos de Gallina Libre',
    categoryId: 'lacteos',
    producerId: 'ana-lucia-mora',
    region: 'Boyacá',
    price: 14000,
    unit: 'bandeja',
    rating: 4.8,
    reviews: 27,
    available: true,
    tags: ['gallinas libres', 'yema intensa'],
    description:
      'Huevos de gallinas criadas en libertad, alimentadas con maíz y forraje natural. Yema de color intenso y cascarón firme, empacados en bandeja de 30 unidades.',
  },
  {
    id: 'rosas-rojas',
    name: 'Rosas Rojas de Altura',
    categoryId: 'flores',
    producerId: 'ana-lucia-mora',
    region: 'Boyacá',
    price: 25000,
    unit: 'docena',
    rating: 4.9,
    reviews: 9,
    available: true,
    tags: ['cultivo de altura', 'tallo largo'],
    description:
      'Rosas rojas cultivadas a 2.600 metros de altura, con tallos largos y botón cerrado. Cortadas el día del despacho para que duren hasta dos semanas en florero.',
  },
  {
    id: 'polen-apicola',
    name: 'Polen Apícola',
    categoryId: 'miel',
    producerId: 'rosa-camargo',
    region: 'Nariño',
    price: 18000,
    unit: '250 g',
    rating: 5.0,
    reviews: 8,
    available: true,
    tags: ['rico en proteína', 'floración nativa'],
    description:
      'Polen apícola recolectado por abejas de floración nativa, rico en proteínas y vitaminas. Deshidratado suavemente para conservar sus nutrientes, en empaque de 250 gramos.',
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