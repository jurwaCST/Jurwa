import type { ReactNode } from 'react'

interface Props {
  title: string
  intro: string
  children: ReactNode
}

/**
 * AuthLayout is the frame for Sign in and Sign up: on a laptop the CST gate
 * and the logo on one side and the form on the other; on a phone one card.
 */
export default function AuthLayout({ title, intro, children }: Props) {
  return (
    <div className="auth">
      <aside className="auth-aside" aria-hidden="true">
        <img src="/campus/gate.jpg" alt="" className="auth-photo" />
        <img
          src="/cst-logo.jpg"
          alt=""
          className="auth-logo"
          width={64}
          height={64}
          onError={(e) => (e.currentTarget.style.display = 'none')}
        />
        <p className="auth-aside-title">Jurwa</p>
        <p className="auth-aside-text">Facility reports for the College of Science and Technology, Rinchending.</p>
      </aside>
      <div className="auth-card">
        <h1>{title}</h1>
        <p className="muted">{intro}</p>
        {children}
      </div>
    </div>
  )
}
