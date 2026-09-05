export default function Logo({ variant = 'full', className, ...props }) {
  if (variant === 'icon') {
    return (
      <img
        src="/assets/images/logo.png"
        alt="Milpa"
        className={className}
        style={{ objectFit: 'cover', objectPosition: 'left', aspectRatio: '1 / 1' }}
        {...props}
      />
    )
  }
  return (
    <img
      src="/assets/images/logo.png"
      alt="Milpa"
      className={className}
      {...props}
    />
  )
}