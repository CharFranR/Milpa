const DEFAULT_MAX_DIM = 1280
const DEFAULT_MAX_BYTES = 200 * 1024

export function byteLength(dataUrl) {
  const comma = dataUrl.indexOf(',')
  if (comma === -1) return dataUrl.length
  return Math.floor((dataUrl.length - comma - 1) * 0.75)
}

function readAsDataURL(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result)
    reader.onerror = () => reject(reader.error)
    reader.readAsDataURL(file)
  })
}

function loadImage(src) {
  return new Promise((resolve, reject) => {
    const img = new Image()
    img.onload = () => resolve(img)
    img.onerror = () => reject(new Error('image decode failed'))
    img.src = src
  })
}

export async function compressImage(file, { maxDim = DEFAULT_MAX_DIM, maxBytes = DEFAULT_MAX_BYTES } = {}) {
  const raw = await readAsDataURL(file)
  try {
    const img = await loadImage(raw)
    const naturalW = img.naturalWidth || img.width
    const naturalH = img.naturalHeight || img.height
    if (byteLength(raw) <= maxBytes && Math.max(naturalW, naturalH) <= maxDim) {
      return raw
    }

    const canvas = document.createElement('canvas')
    const ctx = canvas.getContext('2d')
    if (!ctx) return raw

    const scale = Math.min(1, maxDim / Math.max(naturalW, naturalH))
    let width = Math.max(1, Math.round(naturalW * scale))
    let height = Math.max(1, Math.round(naturalH * scale))

    let quality = 0.82
    const encode = () => {
      canvas.width = width
      canvas.height = height
      ctx.drawImage(img, 0, 0, width, height)
      return canvas.toDataURL('image/jpeg', quality)
    }

    let out = encode()
    while (byteLength(out) > maxBytes && quality > 0.45) {
      quality -= 0.12
      out = encode()
    }
    while (byteLength(out) > maxBytes && Math.max(width, height) > 480) {
      width = Math.round(width * 0.7)
      height = Math.round(height * 0.7)
      quality = 0.7
      out = encode()
      while (byteLength(out) > maxBytes && quality > 0.35) {
        quality -= 0.1
        out = encode()
      }
    }
    return out
  } catch {
    return raw
  }
}
