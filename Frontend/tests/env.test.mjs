import assert from "node:assert/strict";
import test from "node:test";
import { apiBaseUrl, contractDebugEnabled, isDevelopment, normalizeApiBaseUrl } from "../src/config/env.ts";

// This suite runs outside Vite, so `import.meta.env` is absent and every value
// has to fall back to its default. That is the same situation the API tests
// rely on when they assert a request went to "/api/people".
test("outside Vite the configuration falls back to same-origin defaults", () => {
  assert.equal(isDevelopment, false);
  assert.equal(apiBaseUrl, "/api");
  // Unset logging follows the build mode, and this is not a development build.
  assert.equal(contractDebugEnabled, false);
});

test("a configured base URL is trimmed and loses its trailing slashes", () => {
  assert.equal(normalizeApiBaseUrl(undefined), "");
  assert.equal(normalizeApiBaseUrl(""), "");
  assert.equal(normalizeApiBaseUrl("   "), "");
  assert.equal(normalizeApiBaseUrl("/api"), "/api");
  assert.equal(normalizeApiBaseUrl("/api/"), "/api");
  assert.equal(normalizeApiBaseUrl("/api///"), "/api");
  assert.equal(normalizeApiBaseUrl("http://localhost:8080/"), "http://localhost:8080");
  assert.equal(normalizeApiBaseUrl("  http://localhost:8080  "), "http://localhost:8080");
});

// The base URL is a prefix that paths are appended to, so a trailing slash left
// on would turn every request into "//people".
test("a path can be appended without doubling the slash", () => {
  const base = normalizeApiBaseUrl("http://localhost:8080/");
  assert.equal(`${base}/people`, "http://localhost:8080/people");
});
