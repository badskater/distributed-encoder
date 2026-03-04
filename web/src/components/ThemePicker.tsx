import { useTheme, type Theme } from '../theme'

const THEMES: { id: Theme; icon: string; label: string }[] = [
  { id: 'light', icon: '☀', label: 'Light' },
  { id: 'dark',  icon: '●', label: 'Dark'  },
  { id: 'dim',   icon: '◑', label: 'Dim'   },
]

export default function ThemePicker() {
  const { theme, setTheme } = useTheme()
  return (
    <div className="flex items-center gap-0.5">
      {THEMES.map(t => (
        <button
          key={t.id}
          onClick={() => setTheme(t.id)}
          title={t.label}
          className={`w-7 h-7 rounded text-sm transition-colors ${
            theme === t.id
              ? 'bg-th-nav-hover text-white'
              : 'text-th-nav-text hover:bg-th-nav-hover hover:text-white'
          }`}
        >
          {t.icon}
        </button>
      ))}
    </div>
  )
}
