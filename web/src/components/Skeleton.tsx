interface Props {
  /** How many grey lines to show. */
  lines?: number
  /** What is loading, for screen readers ("Loading reports"). */
  label: string
}

/** Skeleton shows grey lines where content will appear while it loads. */
export default function Skeleton({ lines = 3, label }: Props) {
  return (
    <div className="skeleton" role="status" aria-busy="true" aria-label={label}>
      {Array.from({ length: lines }, (_, i) => (
        <span key={i} className="skeleton-line" />
      ))}
    </div>
  )
}
