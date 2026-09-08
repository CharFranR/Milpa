import { cn } from '../../lib/cn'

const SIZE_CLASSES = {
  sm: 'px-4 py-2 text-sm',
  md: 'px-5 py-2.5 text-sm',
  lg: 'px-7 py-3.5 text-base',
}

const VARIANT_CLASSES = {
  primary: 'bg-brand text-white hover:bg-brand-dark',
  accent: 'bg-accent text-night hover:brightness-95',
  outline: 'border-2 border-brand/30 text-brand hover:border-brand hover:bg-brand-soft',
  white: 'border-2 border-white text-white hover:bg-white/10',
  ghost: 'text-brand hover:bg-brand-soft',
  danger: 'bg-red-600 text-white hover:bg-red-700',
  whatsapp: 'bg-whatsapp text-white hover:brightness-95',
  dark: 'bg-night text-white hover:bg-night-soft',
}

export default function Button({
  variant = 'primary',
  size = 'md',
  icon,
  iconRight,
  className,
  type = 'button',
  children,
  ...rest
}) {
  return (
    <button
      type={type}
      className={cn(
        'inline-flex items-center justify-center gap-2 rounded-full font-semibold transition-colors duration-200',
        'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand',
        'disabled:cursor-not-allowed disabled:opacity-50',
        SIZE_CLASSES[size],
        VARIANT_CLASSES[variant],
        className,
      )}
      {...rest}
    >
      {icon}
      {children}
      {iconRight}
    </button>
  )
}