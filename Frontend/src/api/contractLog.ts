import { contractDebugEnabled } from "../config/env.ts";

/**
 * Development-only diagnostics for the API boundary.
 *
 * Every contract call logs its request, response and duration so payloads can
 * be checked against the API contract in DevTools. Whether logging is on comes
 * from `VITE_API_CONTRACT_DEBUG`, which defaults to "on while developing" — see
 * `src/config/env.ts` and `.env`.
 */
type LogContext = Record<string, unknown>;

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
  if (contractDebugEnabled) logRequest(operation, context);

  try {
    const result = await execute();
    if (contractDebugEnabled) {
      logResponse(operation, { ...context, ...resultContext(result) }, result, Math.round(performance.now() - startedAt));
    }
    return result;
  } catch (error) {
    if (contractDebugEnabled) {
      logFailure(operation, context, error, Math.round(performance.now() - startedAt));
    }
    throw error;
  }
}
