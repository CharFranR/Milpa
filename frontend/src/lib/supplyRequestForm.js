import { EMPTY_TIME, toInputValue } from './supplyStatus'

export const EMPTY_FORM = {
  product_name: '',
  total_amount: '',
  unit_of_measure: '0',
  Department: '',
  Municipality: '',
  AddressLine: '',
  request_deadline: '',
  delivery_deadline: '',
  description: '',
  multiple_providers: false,
  min_amount_per_provider: '',
}

export function toForm(dto) {
  return {
    product_name: dto.product_name ?? '',
    total_amount: dto.total_amount ?? '',
    unit_of_measure: String(dto.unit_of_measure ?? 0),
    Department: dto.address?.Department ?? '',
    Municipality: dto.address?.Municipality ?? '',
    AddressLine: dto.address?.AddressLine ?? '',
    request_deadline: toInputValue(dto.request_deadline),
    delivery_deadline: toInputValue(dto.delivery_deadline),
    description: dto.description ?? '',
    multiple_providers: Boolean(dto.multiple_providers),
    min_amount_per_provider: dto.min_amount_per_provider ?? '',
  }
}

function toIso(value, editing) {
  if (value) return new Date(value).toISOString()
  return editing ? EMPTY_TIME : undefined
}

export function validate(form) {
  if (!form.product_name.trim()) return 'Indica el producto que necesitas.'
  if (!(Number(form.total_amount) > 0)) return 'El monto total debe ser mayor que cero.'
  if (!form.Department.trim()) return 'El departamento es obligatorio.'
  if (
    form.request_deadline &&
    form.delivery_deadline &&
    new Date(form.request_deadline) > new Date(form.delivery_deadline)
  ) {
    return 'La fecha de solicitud no puede ser posterior a la fecha de entrega.'
  }
  return ''
}

export function buildPayload(form, base, editing) {
  const payload = {
    // El PATCH del backend pisa las 13 columnas: se reenvía el DTO original
    // completo (incluye actual_amount, address.ID y los importes derivados).
    ...base,
    product_name: form.product_name.trim(),
    total_amount: Number(form.total_amount),
    unit_of_measure: Number(form.unit_of_measure),
    address: {
      ...base?.address,
      Department: form.Department.trim(),
      Municipality: form.Municipality.trim(),
      AddressLine: form.AddressLine.trim(),
    },
    description: form.description,
    multiple_providers: form.multiple_providers,
    min_amount_per_provider:
      form.min_amount_per_provider === '' ? 0 : Number(form.min_amount_per_provider),
  }

  const request_deadline = toIso(form.request_deadline, editing)
  const delivery_deadline = toIso(form.delivery_deadline, editing)
  if (request_deadline !== undefined) payload.request_deadline = request_deadline
  if (delivery_deadline !== undefined) payload.delivery_deadline = delivery_deadline

  if (base) {
    // available = total − ya comprometido con matches; al cambiar el total se
    // conserva lo comprometido y el resto queda disponible.
    const matched = Number(base.total_amount || 0) - Number(base.actual_amount || 0)
    payload.actual_amount = Math.min(
      Math.max(payload.total_amount - matched, 0),
      payload.total_amount,
    )
  }

  return payload
}
