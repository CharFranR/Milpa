const IMAGES_KEY = 'milpa_product_images'

function getImages() {
  try {
    return JSON.parse(localStorage.getItem(IMAGES_KEY) || '{}')
  } catch {
    return {}
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

export function removeProductImage(offeringId) {
  const images = getImages()
  delete images[offeringId]
  localStorage.setItem(IMAGES_KEY, JSON.stringify(images))
}
