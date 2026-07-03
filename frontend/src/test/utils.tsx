import {
  QueryClient,
  QueryClientProvider,
  type QueryClientConfig,
} from '@tanstack/react-query'
import { render, type RenderOptions } from '@testing-library/react'
import type { ReactElement, ReactNode } from 'react'

/**
 * Creates a fresh QueryClient for tests with caching disabled so each test
 * observes deterministic fetch/refetch behaviour.
 *
 * Use this with hooks/components that depend on @tanstack/react-query. Wrap
 * with {@link renderWithQueryClient} for component tests, or pass the wrapper
 * from {@link createQueryWrapper} to renderHook.
 */
export function createTestQueryClient(config?: QueryClientConfig): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        staleTime: 0,
        gcTime: 0,
        refetchOnWindowFocus: false,
      },
      ...config?.defaultOptions,
    },
    ...config,
  })
}

/** A wrapper that provides a QueryClientProvider context. */
export function createQueryWrapper(client?: QueryClient) {
  const queryClient = client ?? createTestQueryClient()
  function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  }
  return { client: queryClient, wrapper: Wrapper }
}

/**
 * Renders a component inside a QueryClientProvider. Returns the standard
 * Testing Library render result plus the QueryClient for assertions.
 */
export function renderWithQueryClient(
  ui: ReactElement,
  options?: RenderOptions & { client?: QueryClient },
) {
  const client = options?.client ?? createTestQueryClient()
  function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={client}>{children}</QueryClientProvider>
  }
  return { ...render(ui, { wrapper: Wrapper, ...options }), client }
}
