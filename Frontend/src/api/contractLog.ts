/**
 * Development-only diagnostics for the API boundary.
 *
 * Every contract call logs its request, response and duration in development so
 * payloads can be checked against the API contract in DevTools.
 */
type LogContext = Record<string, unknown>;

const environment = (import.meta as ImportMeta & {
  env?: { DEV?: boolean; VITE_API_CONTRACT_DEBUG?: string };
}).env;

function loggingEnabled(): boolean {
  return environment?.VITE_API_CONTRACT_DEBUG === "true"
    || (environment?.DEV === true && environment.VITE_API_CONTRACT_DEBUG !== "false");
}

function logRequest(operation: string, context: LogContext): void {
  const { payload, ...metadata } = context;
  console.groupCollapsed(`[api-contract] Request: ${operation}`);
  console.log("Payload sent:", payload ?? {});
  if (Object.keys(metadata).length) console.log("Metadata:", metadata);
  console.groupEnd();
}

function logResponse<T>(operation: string, context: LogContext, result: T, durationMs: number): void {
  const { payload: _payload, ...metadata } = context;
  console.groupCollapsed(`[api-contract] Response: ${operation}`);
  console.log("Duration:", `${durationMs}ms`);
  console.log("Payload received:", result);
  if (Object.keys(metadata).length) console.log("Metadata:", metadata);
  console.groupEnd();
}

function logFailure(operation: string, context: LogContext, error: unknown, durationMs: number): void {
  const { payload, ...metadata } = context;
  console.groupCollapsed(`[api-contract] Failed: ${operation}`);
  console.error("Error:", error);
  console.log("Duration:", `${durationMs}ms`);
  console.log("Payload sent:", payload ?? {});
  if (Object.keys(metadata).length) console.log("Metadata:", metadata);
  console.groupEnd();
}

export async function withContractLog<T>(
  operation: string,
  context: LogContext,
  execute: () => Promise<T>,
  resultContext: (result: T) => LogContext = () => ({}),
): Promise<T> {
  const startedAt = performance.now();
  if (loggingEnabled()) logRequest(operation, context);

  try {
    const result = await execute();
    if (loggingEnabled()) {
      logResponse(operation, { ...context, ...resultContext(result) }, result, Math.round(performance.now() - startedAt));
    }
    return result;
  } catch (error) {
    if (loggingEnabled()) {
      logFailure(operation, context, error, Math.round(performance.now() - startedAt));
    }
    throw error;
  }
}
