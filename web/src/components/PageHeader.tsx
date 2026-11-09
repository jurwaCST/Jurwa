import type { ReactNode } from 'react'

interface Props {
  title: string
  /** One line saying what the page is for. */
  description?: string
  /** Optional: a tag next to the title, or the page's one main action. */
  aside?: ReactNode
}

/** PageHeader is the title block every page starts with. */
export default function PageHeader({ title, description, aside }: Props) {
  return (
    <div className="page-header">
      <div>
        <h1>{title}</h1>
        {description && <p className="muted">{description}</p>}
      </div>
      {aside && <div className="page-header-aside">{aside}</div>}
    </div>
  )
}
