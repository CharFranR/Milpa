const ICONS = {
  frutales: 'local_florist',
  citrusos: 'nutrition',
  citricos: 'nutrition',
  otros: 'shopping_basket',
  cafe: 'coffee',
  granos_basicos: 'grain',
  hortalizas: 'grass',
  lacteos: 'egg',
  carnes: 'kebab_dining',
}

const FALLBACK = 'category'

function normalize(value) {
  return String(value || '')
    .toLowerCase()
    .trim()
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .replace(/[\s-]+/g, '_')
}

export function iconForCategory(category) {
  if (!category) return FALLBACK

  for (const candidate of [category.main_category, category.name]) {
    const hit = ICONS[normalize(candidate)]
    if (hit) return hit
  }

  return category.icon || FALLBACK
}
