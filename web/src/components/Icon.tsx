// Small inline SVG icons (24 x 24, stroke 2), drawn for Jurwa. They take the
// text colour, so they follow the light and dark schemes. Decorative by
// default: give the icon a `label` only when it stands alone.

const paths = {
  home: 'M3 11.5 12 4l9 7.5M5.5 10v9.5h5v-5h3v5h5V10',
  report: 'M12 5v14M5 12h14',
  queue: 'M8 6h12M8 12h12M8 18h12M4 6h.01M4 12h.01M4 18h.01',
  dashboard: 'M4 20V10M10 20V4M16 20v-7M22 20H2',
  admin: 'M12 3l7 3v5c0 4.5-3 8.5-7 10-4-1.5-7-5.5-7-10V6l7-3Z',
  user: 'M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8Zm-7 9a7 7 0 0 1 14 0',
  signout: 'M15 4h3a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2h-3M10 16l-4-4 4-4M6 12h10',
  open: 'M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18Z',
  assigned: 'M15 19a5 5 0 0 0-10 0M10 11a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7ZM17 8l2 2 4-4',
  progress: 'M12 21a9 9 0 1 0 0-18M12 7v5l3 2',
  resolved: 'M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18ZM8 12.5l2.7 2.7L16 10',
  escalated: 'M12 3 2.5 20h19L12 3ZM12 10v4M12 17.5h.01',
  readonly: 'M2.5 12S6 5 12 5s9.5 7 9.5 7-3.5 7-9.5 7-9.5-7-9.5-7ZM12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6Z',
  back: 'M15 18l-6-6 6-6',
  bell: 'M6 9a6 6 0 1 1 12 0c0 6 2.5 7.5 2.5 7.5h-17S6 15 6 9ZM10 20a2 2 0 0 0 4 0',
  star: 'M12 3.5l2.6 5.3 5.9.9-4.3 4.1 1 5.8-5.2-2.8-5.2 2.8 1-5.8-4.3-4.1 5.9-.9L12 3.5Z',
} as const

export type IconName = keyof typeof paths

interface Props {
  name: IconName
  /** Accessible name; leave out when text next to the icon says the same. */
  label?: string
  size?: number
}

export default function Icon({ name, label, size = 20 }: Props) {
  return (
    <svg
      className="icon"
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      role={label ? 'img' : undefined}
      aria-label={label}
      aria-hidden={label ? undefined : true}
      focusable="false"
    >
      <path d={paths[name]} />
    </svg>
  )
}
