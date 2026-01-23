import React from 'react'
import { createRoot } from 'react-dom/client'
import App from './App'
import './i18n'
import './styles/main.css'
import '@fontsource/space-grotesk/400.css'
import '@fontsource/space-grotesk/500.css'
import '@fontsource/space-grotesk/600.css'
import { resolvePalette, resolveTheme } from './theme'

const storedTheme = resolveTheme(window.localStorage.getItem('los-theme'))
document.documentElement.setAttribute('data-theme', storedTheme)

const storedPalette = resolvePalette(window.localStorage.getItem('los-palette'))
document.documentElement.setAttribute('data-palette', storedPalette)

const root = document.getElementById('root')
if (root) {
  createRoot(root).render(
    <React.StrictMode>
      <App />
    </React.StrictMode>
  )
}
