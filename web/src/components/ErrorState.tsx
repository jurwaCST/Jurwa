import { LostArt } from './Art'

interface Props {
  /** What went wrong, from ApiError.message or a plain sentence. */
  message: string
  /** Shown as a "Try again" button when given. */
  onRetry?: () => void
  title?: string
}

/** ErrorState replaces content that could not load, with a way to try again. */
export default function ErrorState({ message, onRetry, title = 'This did not load' }: Props) {
  return (
    <div className="empty error-state" role="alert">
      <div className="empty-art">
        <LostArt />
      </div>
      <h2>{title}</h2>
      <p>{message}</p>
      {onRetry && (
        <button type="button" className="btn btn-secondary" onClick={onRetry}>
          Try again
        </button>
      )}
    </div>
  )
}
