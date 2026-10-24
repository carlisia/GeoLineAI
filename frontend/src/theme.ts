import { createTheme } from '@mui/material/styles'

/**
 * GeoLineAI's palette, from the author's site: teal for the controls a person acts
 * with, violet for anything a model wrote or a model action. Every text colour here
 * meets WCAG AA (4.5:1) on white; the light border is decorative only, and inputs
 * use the darker one so their outline meets 3:1.
 */
export const colors = {
  primary: '#487f89',
  primaryDark: '#36646c',
  model: '#825ce5',
  modelDark: '#6a45cc',
  text: '#2b2b2b',
  textSecondary: '#707070',
  surface: '#f4f4f4',
  panel: '#ffffff',
  border: '#b8b8b8',
  inputBorder: '#707070',
  pass: '#2e7d32',
  fail: '#c62828',
  review: '#8a5a00',
}

export const fonts = {
  heading: 'Inter, system-ui, -apple-system, sans-serif',
  body: '"IBM Plex Mono", ui-monospace, Menlo, monospace',
  code: '"Fira Code", ui-monospace, Menlo, monospace',
}

/** A button that starts a model action, filled in the model violet. */
export const modelButton = {
  bgcolor: colors.model,
  color: colors.panel,
  '&:hover': { bgcolor: colors.modelDark },
}

/** A white panel on the page's surface. */
export const panel = {
  bgcolor: colors.panel,
  border: `1px solid ${colors.border}`,
  borderRadius: 1,
  p: { xs: 2, sm: 2.5 },
}

export const theme = createTheme({
  palette: {
    mode: 'light',
    primary: { main: colors.primary, dark: colors.primaryDark, contrastText: colors.panel },
    secondary: { main: colors.model, dark: colors.modelDark, contrastText: colors.panel },
    error: { main: colors.fail },
    success: { main: colors.pass },
    warning: { main: colors.review },
    text: { primary: colors.text, secondary: colors.textSecondary },
    background: { default: colors.surface, paper: colors.panel },
    divider: colors.border,
  },
  typography: {
    fontFamily: fonts.body,
    fontSize: 13,
    h1: { fontFamily: fonts.heading, fontSize: 24, fontWeight: 700 },
    h2: { fontFamily: fonts.heading, fontSize: 18, fontWeight: 600 },
    h3: { fontFamily: fonts.heading, fontSize: 15, fontWeight: 600 },
    overline: { fontFamily: fonts.heading, fontWeight: 600, letterSpacing: '0.06em', lineHeight: 1.8 },
    button: { fontFamily: fonts.heading, fontWeight: 600, textTransform: 'none' },
  },
  shape: { borderRadius: 6 },
  components: {
    MuiOutlinedInput: { styleOverrides: { notchedOutline: { borderColor: colors.inputBorder } } },
    MuiButton: { defaultProps: { disableElevation: true } },
  },
})
