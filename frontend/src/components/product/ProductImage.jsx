import { useState } from 'react'
import Logo from '../Logo'
import { cn } from '../../lib/cn'

export default function ProductImage({ image_url, name, className, imgClassName }) {
  const [hasError, setHasError] = useState(false)

  const src = image_url || ''

  return (
    <div
      className={cn(
        'relative overflow-hidden bg-brand-soft',
        className,
      )}
    >
      {src && !hasError ? (
        <img
          src={src}
          alt={name}
          loading="lazy"
          onError={() => setHasError(true)}
          className={cn('h-full w-full object-cover', imgClassName)}
        />
      ) : (
        <span
          aria-hidden="true"
          className="flex h-full w-full items-center justify-center text-brand/30"
        >
          <Logo variant="icon" className="w-12 h-12" />
        </span>
      )}
    </div>
  )
}
