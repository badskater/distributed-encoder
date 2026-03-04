/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        th: {
          bg:               'var(--th-bg)',
          surface:          'var(--th-surface)',
          'surface-muted':  'var(--th-surface-muted)',
          text:             'var(--th-text)',
          'text-secondary': 'var(--th-text-secondary)',
          'text-muted':     'var(--th-text-muted)',
          'text-subtle':    'var(--th-text-subtle)',
          border:           'var(--th-border)',
          'border-subtle':  'var(--th-border-subtle)',
          'nav-bg':         'var(--th-nav-bg)',
          'nav-active':     'var(--th-nav-active)',
          'nav-hover':      'var(--th-nav-hover)',
          'nav-text':       'var(--th-nav-text)',
          'input-bg':       'var(--th-input-bg)',
          'input-border':   'var(--th-input-border)',
          'log-bg':         'var(--th-log-bg)',
          'log-text':       'var(--th-log-text)',
          'progress-track': 'var(--th-progress-track)',
        },
      },
    },
  },
  plugins: [],
}
