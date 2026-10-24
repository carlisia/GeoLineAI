import { useEffect } from 'react'

/**
 * While `active`, asks the browser to warn before a reload or page exit: a chosen or
 * saving verdict would be lost without a word.
 */
export function useUnloadWarning(active: boolean) {
  useEffect(() => {
    if (!active) return
    const warn = (event: BeforeUnloadEvent) => event.preventDefault()
    window.addEventListener('beforeunload', warn)
    return () => window.removeEventListener('beforeunload', warn)
  }, [active])
}
