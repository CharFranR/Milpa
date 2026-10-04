import { cn } from '../../lib/cn'

export default function Skeleton({ className, circle = false }) {
  return (
    <div
      aria-hidden="true"
      className={cn('animate-pulse rounded-xl bg-gray-100', circle && 'rounded-full', className)}
    />
  )
}
