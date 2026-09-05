const IMAGE_MARKER = '\n\nImageBase64:'
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
  return `${description}${IMAGE_MARKER}${imageUrl}`
}

export function extractImageFromDescription(description) {
  if (!description) return { clean: '', imageUrl: '' }
  const idx = description.indexOf(IMAGE_MARKER)
  if (idx === -1) return { clean: description, imageUrl: '' }
  return {
    clean: description.slice(0, idx),
    imageUrl: description.slice(idx + IMAGE_MARKER.length),
  }
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
