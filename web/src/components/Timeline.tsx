import Icon from './Icon'
import { statusLooks, type Status } from './labels'

export interface TimelineEntry {
  status: Status
  /** ISO time of the change, for example from report-history. */
  at: string
  /** Who made the change, for example "DSA" or a staff name. */
  by?: string
  note?: string
}

// "8 Nov 2026, 14:05" in the reader's own time zone.
const when = new Intl.DateTimeFormat('en-GB', { dateStyle: 'medium', timeStyle: 'short' })

/** Timeline shows a report's status changes, in the order given (oldest first). */
export default function Timeline({ entries }: { entries: TimelineEntry[] }) {
  if (entries.length === 0) {
    return <p className="muted">No changes yet.</p>
  }
  return (
    <ol className="timeline">
      {entries.map((e, i) => {
        const look = statusLooks[e.status]
        return (
          <li key={`${e.status}-${e.at}-${i}`} className={`timeline-item tone-${look.tone}`}>
            <span className="timeline-dot">
              <Icon name={look.icon} size={16} />
            </span>
            <div className="timeline-body">
              <strong>{look.label}</strong>
              {e.by && <span className="muted"> by {e.by}</span>}
              <time dateTime={e.at} className="timeline-time">
                {when.format(new Date(e.at))}
              </time>
              {e.note && <p className="timeline-note">{e.note}</p>}
            </div>
          </li>
        )
      })}
    </ol>
  )
}
