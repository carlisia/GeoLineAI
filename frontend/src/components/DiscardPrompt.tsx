import Button from '@mui/material/Button'
import Dialog from '@mui/material/Dialog'
import DialogActions from '@mui/material/DialogActions'
import DialogContent from '@mui/material/DialogContent'
import DialogContentText from '@mui/material/DialogContentText'
import DialogTitle from '@mui/material/DialogTitle'

interface DiscardPromptProps {
  open: boolean
  onDiscard: () => void
  onKeep: () => void
}

/** Asks before changing scenarios abandons a chosen verdict. Dismissing the prompt keeps it. */
export function DiscardPrompt({ open, onDiscard, onKeep }: DiscardPromptProps) {
  return (
    <Dialog open={open} onClose={onKeep}>
      <DialogTitle>Discard unsaved verdict?</DialogTitle>
      <DialogContent>
        <DialogContentText>The verdict you chose for this reading has not been saved. Discard it, or keep it and save it first.</DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button variant="outlined" onClick={onDiscard}>
          Discard
        </Button>
        <Button variant="contained" autoFocus onClick={onKeep}>
          Keep editing
        </Button>
      </DialogActions>
    </Dialog>
  )
}
