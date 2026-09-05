import { useState } from 'react'
import Icon from '../ui/Icon'
import { productImageUrl } from '../../mocks/catalog'
import { cn } from '../../lib/cn'

export default function ProductImage({ productId, name, className, imgClassName }) {
  const [hasError, setHasError] = useState(false)

  return (
    <div
      className={cn(
        'relative overflow-hidden bg-brand-soft',
        className,
      )}
    >
      {!hasError ? (
        <img
          src={productImageUrl(productId)}
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
          <Icon name="eco" size={48} />
        </span>
      )}
    </div>
  )
}
