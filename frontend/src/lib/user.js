export function getDisplayName(user) {
  if (!user) return 'Usuario'
  const first = user.first_name || ''
  const last = user.last_name || ''
  const full = `${first} ${last}`.trim()
  return full || 'Usuario'
}

export function getInitials(user) {
  if (!user) return '??'
  const first = user.first_name || ''
  const last = user.last_name || ''
  if (first && last) return (first[0] + last[0]).toUpperCase()
  if (first) return first.slice(0, 2).toUpperCase()
  if (last) return last.slice(0, 2).toUpperCase()
  return '??'
}
