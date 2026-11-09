import type { ButtonHTMLAttributes } from 'react'

type Variant = 'primary' | 'secondary' | 'ghost' | 'danger'

interface Props extends ButtonHTMLAttributes<HTMLButtonElement> {
  /** primary: the one saffron action of a screen. Default: primary. */
  variant?: Variant
  /** While true the button is disabled, busy and shows `loadingText`. */
  loading?: boolean
  loadingText?: string
}

/** Button with the app's four styles and a loading state. */
export default function Button({
  variant = 'primary',
  loading = false,
  loadingText,
  className,
  disabled,
  type = 'button',
  children,
  ...rest
}: Props) {
  const classes = ['btn', `btn-${variant}`, loading ? 'is-loading' : null, className].filter(Boolean).join(' ')
  return (
    <button {...rest} type={type} className={classes} disabled={disabled || loading} aria-busy={loading || undefined}>
      {loading && <span className="btn-spinner" aria-hidden="true" />}
      {loading && loadingText ? loadingText : children}
    </button>
  )
}
