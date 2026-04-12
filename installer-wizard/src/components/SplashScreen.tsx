import { useEffect, useState } from 'react'

interface SplashScreenProps {
  onComplete: () => void
  appTitle?: string
  subtitle?: string
}

const FIRST_LAUNCH_KEY = 'catalogizer_launched'
const FIRST_LAUNCH_DURATION = 5000
const REGULAR_LAUNCH_DURATION = 2500

export function SplashScreen({
  onComplete,
  appTitle = 'Catalogizer',
  subtitle = 'Your media collection, organized',
}: SplashScreenProps) {
  const [isFirstLaunch] = useState(() => {
    return localStorage.getItem(FIRST_LAUNCH_KEY) === null
  })

  useEffect(() => {
    const duration = isFirstLaunch ? FIRST_LAUNCH_DURATION : REGULAR_LAUNCH_DURATION

    localStorage.setItem(FIRST_LAUNCH_KEY, 'true')

    const timer = setTimeout(() => {
      onComplete()
    }, duration)

    return () => clearTimeout(timer)
  }, [isFirstLaunch, onComplete])

  return (
    <div
      className="fixed inset-0 z-50 flex flex-col items-center justify-center"
      style={{
        background: 'linear-gradient(to bottom, #0F172A, #1E293B)',
      }}
    >
      <div className="flex flex-col items-center">
        <div
          className="w-28 h-28 rounded-2xl shadow-2xl border-[3px] mb-6 flex items-center justify-center"
          style={{
            background: 'linear-gradient(135deg, #1E293B 0%, #0F172A 100%)',
            borderColor: '#DC2626',
          }}
          aria-label={`${appTitle} icon`}
        >
          <span className="text-6xl font-extrabold text-white select-none">C</span>
        </div>

        <h1 className="text-3xl font-bold text-white mb-2">{appTitle}</h1>

        <p className="text-slate-400 text-sm mb-8">{subtitle}</p>

        <div className="w-8 h-8 border-[3px] border-blue-500 border-t-transparent rounded-full animate-spin" />
      </div>

      <div className="absolute bottom-8 flex flex-col items-center gap-3">
        <span className="text-slate-500 text-xs">
          Made with &#9829; by Vasic Digital
        </span>
        <span className="text-slate-600 text-xs">v1.1.0</span>
      </div>
    </div>
  )
}
