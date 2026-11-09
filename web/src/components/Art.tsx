// Inline SVG illustrations, drawn for Jurwa in the indigo, marigold and CST
// blue palette. Decorative: they are hidden from screen readers.

/** A quiet empty courtyard with a signpost, for 404, 403 and empty lists. */
export function LostArt() {
  return (
    <svg viewBox="0 0 260 180" className="art" aria-hidden="true" focusable="false">
      <ellipse cx="130" cy="160" rx="110" ry="10" fill="currentColor" opacity="0.07" />
      <rect x="124" y="60" width="10" height="100" rx="3" fill="var(--pine-800)" />
      <path d="M64 52h78l14 14-14 14H64Z" fill="var(--saffron-500)" />
      <path d="M196 96h-74l-14 14 14 14h74Z" fill="var(--teal-500)" />
      <circle cx="129" cy="56" r="7" fill="var(--pine-800)" />
    </svg>
  )
}
