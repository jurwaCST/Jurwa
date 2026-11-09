import type { ReactNode } from 'react'

interface Props {
  title: string
  children?: ReactNode
  /** An illustration (inline SVG) shown above the title. */
  art?: ReactNode
}

/** EmptyState is shown when a list is empty, a page is missing or not allowed. */
export default function EmptyState({ title, children, art }: Props) {
  return (
    <div className="empty">
      {art && <div className="empty-art">{art}</div>}
      <h2>{title}</h2>
      {children}
    </div>
  )
}
