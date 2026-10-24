import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterAll, afterEach, beforeAll, expect } from 'vitest'
import { server, strayRequests } from './fakeBackend'

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }))
afterEach(() => {
  cleanup()
  server.resetHandlers()
  // A request the test gave no answer for got a 500, which a failure-path test could
  // pass on for the wrong reason; the test fails instead.
  expect(strayRequests.splice(0)).toEqual([])
})
afterAll(() => server.close())
