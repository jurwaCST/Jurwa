import { useId, useState } from 'react'

interface Props {
  value: string
  onChange: (value: string) => void
  autoComplete: 'current-password' | 'new-password'
  /** A rule shown under the field, for example the minimum length. */
  hint?: string
}

/**
 * PasswordField is a password input with a Show / Hide button. The button sits
 * outside the <label>, so the field's accessible name stays "Password".
 */
export default function PasswordField({ value, onChange, autoComplete, hint }: Props) {
  const [shown, setShown] = useState(false)
  const hintId = useId()

  return (
    <div className="password-field">
      <label>
        Password
        <input
          type={shown ? 'text' : 'password'}
          autoComplete={autoComplete}
          required
          value={value}
          onChange={(e) => onChange(e.target.value)}
          aria-describedby={hint ? hintId : undefined}
        />
      </label>
      <button
        type="button"
        className="password-toggle"
        aria-pressed={shown}
        aria-label={shown ? 'Hide password' : 'Show password'}
        onClick={() => setShown((s) => !s)}
      >
        {shown ? 'Hide' : 'Show'}
      </button>
      {hint && (
        <p id={hintId} className="field-hint">
          {hint}
        </p>
      )}
    </div>
  )
}
