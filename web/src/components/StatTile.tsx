import Icon, { type IconName } from './Icon'

interface Props {
  label: string
  value: string
  icon: IconName
  /** Colour key, for example "escalated"; the label and icon still say it. */
  tone: string
}

/** StatTile is one big number on a dashboard. */
export default function StatTile({ label, value, icon, tone }: Props) {
  return (
    <div className={`stat stat-${tone}`}>
      <span className="stat-icon">
        <Icon name={icon} size={22} />
      </span>
      <span className="stat-label">{label}</span>
      <span className="stat-value num">{value}</span>
    </div>
  )
}
