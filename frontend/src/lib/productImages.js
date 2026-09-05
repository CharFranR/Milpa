const IMAGES_KEY = 'milpa_product_images'

function getImages() {
  try {
    return JSON.parse(localStorage.getItem(IMAGES_KEY) || '{}')
  } catch {
    return {}
  }
}

export function embedImageInDescription(description, imageUrl) {
  if (!imageUrl) return description
  return `${description}\n\nImageBase64:${imageUrl}`
}

export function extractImageFromDescription(description) {
  if (!description) return { clean: '', imageUrl: '' }
  const marker = 'ImageBase64:'
  const idx = description.indexOf(marker)
  if (idx === -1) return { clean: description, imageUrl: '' }
  const before = description.slice(0, idx).replace(/\s+$/, '')
  const after = description.slice(idx + marker.length)
  return { clean: before, imageUrl: after }
}

export function getProductImage(offeringId) {
  const images = getImages()
  return images[offeringId] || null
}

export function setProductImage(offeringId, dataUrl) {
  const images = getImages()
  images[offeringId] = dataUrl
  localStorage.setItem(IMAGES_KEY, JSON.stringify(images))
}

export function resolveOfferingImage(offering) {
  if (!offering) return ''
  const { imageUrl } = extractImageFromDescription(offering.description)
  return imageUrl || offering.image_url || getProductImage(offering.id) || ''
}

export function removeProductImage(offeringId) {
  const images = getImages()
  delete images[offeringId]
  localStorage.setItem(IMAGES_KEY, JSON.stringify(images))
}
