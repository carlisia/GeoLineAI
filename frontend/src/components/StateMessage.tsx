import Button from '@mui/material/Button'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'

interface StateMessageProps {
  kind: 'loading' | 'error'
  title: string
  /** Offers a button when given. */
  action?: { label: string; run: () => void }
}

/** A loading or error state in place of content. */
export function StateMessage({ kind, title, action }: StateMessageProps) {
  return (
    <Stack role={kind === 'error' ? 'alert' : 'status'} sx={{ py: 4, alignItems: 'center', textAlign: 'center', gap: 1.5 }}>
      <Typography variant="body1" color={kind === 'error' ? 'error' : 'textSecondary'}>
        {title}
      </Typography>
      {action && (
        <Button variant="contained" onClick={action.run}>
          {action.label}
        </Button>
      )}
    </Stack>
  )
}
