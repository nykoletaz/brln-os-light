/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      fontFamily: {
        display: ['"Space Grotesk"', 'sans-serif']
      },
      colors: {
        ink: 'rgb(var(--color-ink) / <alpha-value>)',
        slate: 'rgb(var(--color-slate) / <alpha-value>)',
        fog: 'rgb(var(--color-fog) / <alpha-value>)',
        ember: 'rgb(var(--color-ember) / <alpha-value>)',
        glow: 'rgb(var(--color-glow) / <alpha-value>)',
        midnight: 'rgb(var(--color-midnight) / <alpha-value>)',
        brass: 'rgb(var(--color-brass) / <alpha-value>)'
      },
      boxShadow: {
        panel: 'var(--shadow-panel)'
      }
    }
  },
  plugins: []
}
