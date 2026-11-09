import { useId } from 'react'
import Field from './Field'

interface Props {
  label: string
  value: string
  onChange: (value: string) => void
  /** The most characters allowed, for example 500 for a report description. */
  max: number
  hint?: string
  error?: string | null
  rows?: number
  required?: boolean
}

/** Counts characters (Dzongkha letters count as one each), not bytes. */
function length(text: string): number {
  return [...text].length
}

/**
 * TextArea is a labelled multi-line field with a live "12 / 500" counter.
 * Over the limit the counter turns red and the field is marked invalid.
 */
export default function TextArea({ label, value, onChange, max, hint, error, rows = 5, required }: Props) {
  const counterId = useId()
  const count = length(value)
  const over = count > max

  return (
    <Field
      label={label}
      hint={hint}
      error={error}
      after={
        <p id={counterId} className={over ? 'field-counter over' : 'field-counter'} aria-live="polite">
          {count} / {max}
        </p>
      }
    >
      <textarea
        rows={rows}
        required={required}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        aria-invalid={over || error ? true : undefined}
      />
    </Field>
  )
}
