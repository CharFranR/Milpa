import { describe, expect, it } from 'vitest'
import {
  embedImageInDescription,
  extractImageFromDescription,
  getProductImage,
  removeProductImage,
  resolveOfferingImage,
  setProductImage,
} from './productImages'

describe('lib/productImages', () => {
  it('incrusta la imagen detrás del marcador ImageBase64', () => {
    const out = embedImageInDescription('Tomates rojos', 'data:image/png;base64,AAA')
    expect(out).toBe('Tomates rojos\n\nImageBase64:data:image/png;base64,AAA')
    expect(embedImageInDescription('Tomates rojos', '')).toBe('Tomates rojos')
  })

  it('separa la descripción de la imagen incrustada', () => {
    const { clean, imageUrl } = extractImageFromDescription('Tomates rojos\n\nImageBase64:data:image/png;base64,AAA')
    expect(clean).toBe('Tomates rojos')
    expect(imageUrl).toBe('data:image/png;base64,AAA')
  })

  it('devuelve la descripción entera cuando no hay marcador', () => {
    expect(extractImageFromDescription('Solo texto')).toEqual({ clean: 'Solo texto', imageUrl: '' })
    expect(extractImageFromDescription('')).toEqual({ clean: '', imageUrl: '' })
  })

  it('guarda y borra la imagen de un producto en localStorage', () => {
    expect(getProductImage('p1')).toBeNull()
    setProductImage('p1', 'data:image/png;base64,BBB')
    expect(getProductImage('p1')).toBe('data:image/png;base64,BBB')
    removeProductImage('p1')
    expect(getProductImage('p1')).toBeNull()
  })

  it('resuelve la imagen por orden: descripción, image_url y storage', () => {
    expect(resolveOfferingImage(null)).toBe('')
    expect(resolveOfferingImage({ description: 'x', image_url: '/img/a.png' })).toBe('/img/a.png')
    expect(resolveOfferingImage({ description: 'x\nImageBase64:/img/b.png', image_url: '/img/a.png' })).toBe('/img/b.png')

    setProductImage('p9', '/img/c.png')
    expect(resolveOfferingImage({ id: 'p9', description: 'x' })).toBe('/img/c.png')
    expect(removeProductImage('p9')).toBeUndefined()
  })
})
