import { cloneElement, isValidElement, useId, type ReactElement, type ReactNode } from 'react'

interface Props {
  label: string
  /** A rule or example under the field, read with the field by screen readers. */
  hint?: string
  /** What is wrong and how to fix it; marks the control invalid and is announced. */
  error?: string | null
  /** One control: an <input>, <select> or <textarea>. */
  children: ReactElement<Record<string, unknown>>
  /** Extra content under the control, for example a counter. */
  after?: ReactNode
}

/**
 * Field wraps one form control with its label, hint and error. The <label>
 * wraps the control, so the control's name is exactly `label`; the hint and
 * error are linked with aria-describedby.
 */
export default function Field({ label, hint, error, children, after }: Props) {
  const id = useId()
  const hintId = `${id}-hint`
  const errorId = `${id}-error`
  const describedBy = [hint ? hintId : null, error ? errorId : null].filter(Boolean).join(' ') || undefined

  const control = isValidElement(children)
    ? cloneElement(children, {
        'aria-describedby': describedBy,
        // Keep a control's own invalid mark (for example a counter over its limit).
        'aria-invalid': error ? true : children.props['aria-invalid'],
      })
    : children

  return (
    <div className={error ? 'field field-invalid' : 'field'}>
      <label>
        {label}
        {control}
      </label>
      {after}
      {hint && (
        <p id={hintId} className="field-hint">
          {hint}
        </p>
      )}
      {error && (
        <p id={errorId} role="alert" className="field-error">
          {error}
        </p>
      )}
    </div>
  )
}
