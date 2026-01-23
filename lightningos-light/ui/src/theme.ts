export type ThemeMode = 'dark' | 'light'

export type PaletteKey =
  | 'teal'
  | 'ocean'
  | 'sunset'
  | 'orchid'
  | 'forest'
  | 'aurora'
  | 'ember'
  | 'slate'

export const paletteOrder: PaletteKey[] = [
  'teal',
  'ocean',
  'sunset',
  'orchid',
  'forest',
  'aurora',
  'ember',
  'slate'
]

export const defaultPalette: PaletteKey = 'teal'

export const resolveTheme = (value: string | null): ThemeMode => (value === 'light' ? 'light' : 'dark')

export const resolvePalette = (value: string | null): PaletteKey => {
  if (value && paletteOrder.includes(value as PaletteKey)) {
    return value as PaletteKey
  }
  return defaultPalette
}
