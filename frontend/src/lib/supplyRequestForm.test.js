import { describe, expect, it } from 'vitest'
import { EMPTY_FORM, buildPayload, toForm, validate } from './supplyRequestForm'

const DTO = {
  id: 'req-1',
  buyer_id: 'buyer-1',
  product_name: 'Urea',
  total_amount: 25000,
  actual_amount: 25000,
  amount_unit: 0,
  number_of_units: 0,
  amount_per_unit: 0,
  unit_of_measure: 1,
  address: {
    ID: 'addr-9',
    Department: 'Managua',
    Municipality: 'Tipitapa',
    AddressLine: 'Km 12',
    Latitude: 0,
    Longitude: 0,
  },
  request_deadline: '2026-11-15T00:00:00.000Z',
  delivery_deadline: '0001-01-01T00:00:00Z',
  description: 'Temporada de maíz',
  multiple_providers: true,
  min_amount_per_provider: 5000,
  status: 0,
}

describe('lib/supplyRequestForm', () => {
  it('traduce el DTO del backend a valores de formulario', () => {
    const form = toForm(DTO)

    expect(form.product_name).toBe('Urea')
    expect(form.unit_of_measure).toBe('1')
    expect(form.Department).toBe('Managua')
    expect(form.request_deadline).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$/)
    expect(form.delivery_deadline).toBe('')
    expect(form.multiple_providers).toBe(true)
  })

  it('al editar reenvía el DTO completo con el id de la dirección', () => {
    const form = { ...toForm(DTO), product_name: '  Urea mejorada  ', total_amount: '30000' }
    const payload = buildPayload(form, DTO, true)

    expect(payload.product_name).toBe('Urea mejorada')
    expect(payload.address.ID).toBe('addr-9')
    expect(payload.address.Department).toBe('Managua')
    expect(payload.amount_unit).toBe(0)
    expect(payload.number_of_units).toBe(0)
    expect(payload.min_amount_per_provider).toBe(5000)
    expect(payload.status).toBe(0)
  })

  it('conserva lo comprometido con matches al subir el total', () => {
    const base = { ...DTO, total_amount: 25000, actual_amount: 15000 }
    const form = { ...toForm(base), total_amount: '40000' }
    const payload = buildPayload(form, base, true)

    expect(payload.total_amount).toBe(40000)
    expect(payload.actual_amount).toBe(30000)
  })

  it('nunca deja el disponible por debajo de cero ni por encima del total', () => {
    const base = { ...DTO, total_amount: 25000, actual_amount: 5000 }
    const form = { ...toForm(base), total_amount: '8000' }
    const payload = buildPayload(form, base, true)

    expect(payload.actual_amount).toBe(0)
  })

  it('al crear no envía actual_amount ni campos del backend', () => {
    const form = { ...EMPTY_FORM, product_name: 'Arroz', total_amount: '1000', Department: 'León' }
    const payload = buildPayload(form, null, false)

    expect(payload.actual_amount).toBeUndefined()
    expect(payload.id).toBeUndefined()
    expect(payload.status).toBeUndefined()
    expect(payload.total_amount).toBe(1000)
  })

  it('valida campos obligatorios y plazos invertidos', () => {
    expect(validate(EMPTY_FORM)).toContain('producto')
    expect(validate({ ...EMPTY_FORM, product_name: 'X', total_amount: '0', Department: 'M' })).toContain('monto')
    expect(validate({ ...EMPTY_FORM, product_name: 'X', total_amount: '10', Department: '' })).toContain('departamento')

    const inverted = {
      ...EMPTY_FORM,
      product_name: 'X',
      total_amount: '10',
      Department: 'M',
      request_deadline: '2026-12-30T10:00',
      delivery_deadline: '2026-12-01T10:00',
    }
    expect(validate(inverted)).toContain('posterior')
    expect(validate({ ...inverted, request_deadline: '2026-12-01T10:00', delivery_deadline: '2026-12-30T10:00' })).toBe('')
  })
})
