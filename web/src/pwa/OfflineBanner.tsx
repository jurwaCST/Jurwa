// Offline banner (SCRUM-98). Owner: Nyendrak.
// While the phone has no connection, says that the pages shown are the saved
// ones. Uses Gayley's tokens, so it follows the light and dark theme.
import { useEffect, useState } from 'react'
import './OfflineBanner.css'

export default function OfflineBanner() {
  const [offline, setOffline] = useState(() => typeof navigator !== 'undefined' && navigator.onLine === false)

  useEffect(() => {
    const goOffline = () => setOffline(true)
    const goOnline = () => setOffline(false)
    window.addEventListener('offline', goOffline)
    window.addEventListener('online', goOnline)
    return () => {
      window.removeEventListener('offline', goOffline)
      window.removeEventListener('online', goOnline)
    }
  }, [])

  if (!offline) return null
  return (
    <p role="status" className="offline-banner">
      You are offline. Showing saved pages.
    </p>
  )
}
