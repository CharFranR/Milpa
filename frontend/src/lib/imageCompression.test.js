import { describe, expect, it } from 'vitest'
import { byteLength } from './imageCompression'

describe('byteLength', () => {
  it('ignores the data URL header and counts base64 bytes', () => {
    const oneKbBase64 = 'A'.repeat(1024)
    expect(byteLength(`data:image/jpeg;base64,${oneKbBase64}`)).toBe(Math.floor(1024 * 0.75))
  })

  it('falls back to full length without a comma', () => {
    expect(byteLength('data-url-without-comma')).toBe('data-url-without-comma'.length)
  })

  it('approximates real payload size for embedding budget checks', () => {
    const payload = 'B'.repeat(400)
    const dataUrl = `data:image/jpeg;base64,${payload}`
    expect(byteLength(dataUrl)).toBeLessThan(dataUrl.length)
    expect(byteLength(dataUrl)).toBe(300)
  })
})
