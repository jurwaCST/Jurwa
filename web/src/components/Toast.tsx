import { useEffect } from 'react'
import Icon from './Icon'

interface Props {
  /** What just happened, in the words of the button that did it ("Status updated"). */
  message: string
  /** Called when the toast has been shown long enough; remove it then. */
  onDone: () => void
  /** How long it stays, in milliseconds. */
  duration?: number
}

/** Toast confirms an action for a few seconds. Screen readers hear it politely. */
export default function Toast({ message, onDone, duration = 4000 }: Props) {
  useEffect(() => {
    const t = setTimeout(onDone, duration)
    return () => clearTimeout(t)
  }, [onDone, duration])

  return (
    <div className="toast" role="status" aria-live="polite">
      <Icon name="resolved" size={20} />
      {message}
    </div>
  )
}
