import Icon from './Icon'
import { statusLooks, type Status } from './labels'

/** StatusBadge is the pill used wherever a report status is shown. */
export default function StatusBadge({ status }: { status: Status }) {
  const look = statusLooks[status]
  return (
    <span className={`badge badge-${look.tone}`}>
      <Icon name={look.icon} size={16} />
      {look.label}
    </span>
  )
}
