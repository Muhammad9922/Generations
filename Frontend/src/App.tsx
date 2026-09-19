import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Link, Route, Routes, useNavigate } from "react-router";
import PersonDetails from "./pages/PersonDetails";
import { Theme } from "@radix-ui/themes";
import { KBarProvider, type Action } from "kbar";
import { Home as HomeIcon, Moon } from "lucide-react";
import Home from "./pages/Home";
import NewFamily from "./pages/NewFamily";
import CommandPalette from "./components/CommandPalette";
import { createPersonActions } from "./helpers/PersonActions.ts";
import { getAllPeople, type Person } from "./api/people.ts";
import { isAbort } from "./api/http.ts";

/**
 * Owns the shared people list and the session's theme preference. Keeping the
 * Kbar provider above Routes makes search available on both home and details
 * pages. The details page asks for a refresh after it creates or renames
 * someone, so the palette never shows a stale label.
 */
export default function App() {
  // BrowserRouter in the entry point supplies client-side navigation.
  const navigate = useNavigate();
  // One load state feeds both search feedback and the details page.
  const [people, setPeople] = useState<Person[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  // Incrementing this re-reads the list: a failed load retries, and a successful
  // write refreshes the palette.
  const [attempt, setAttempt] = useState(0);
  // Only the first load shows loading feedback; refreshes happen quietly.
  const loadedOnce = useRef(false);
  // Theme is intentionally in-memory: refreshing restores the light default.
  const [dark, setDark] = useState(false);

  useEffect(() => {
    // An abandoned request is cancelled rather than merely ignored, so a slow
    // response cannot overwrite a newer list.
    const controller = new AbortController();
    if (!loadedOnce.current) setLoading(true);
    getAllPeople(controller.signal).then((data) => {
      loadedOnce.current = true;
      setPeople(data);
      setError(null);
    }).catch((cause) => {
      if (isAbort(cause)) return;
      setError(cause instanceof Error ? cause.message : "Could not load people.");
    }).finally(() => {
      if (!controller.signal.aborted) setLoading(false);
    });
    return () => controller.abort();
  }, [attempt]);

  /** Re-reads the list without a page reload; the details page calls this. */
  const refreshPeople = useCallback(() => setAttempt((value) => value + 1), []);

  // Stable action references avoid needless unregister/register cycles. Rebuild
  // when people or theme change so callbacks and the theme label stay current.
  // Person URLs encode IDs, preserving IDs containing reserved URL characters.
  const actions = useMemo<Action[]>(() => [
    { id: "home", name: "Go home", section: "Navigation", keywords: "welcome start", icon: <HomeIcon size={20} />, perform: () => navigate("/") },
    { id: "theme", name: dark ? "Switch to light mode" : "Switch to dark mode", section: "Preferences", keywords: "theme appearance", icon: <Moon size={20} />, perform: () => setDark((value) => !value) },
    ...createPersonActions(people, (person) => navigate(`/people/${encodeURIComponent(person.id)}`)),
  ], [people, navigate, dark]);

  return (
    <Theme appearance={dark ? "dark" : "light"}>
      {/* Kbar supplies its default Ctrl/Cmd+K shortcut and shared search state.
          Reset retry feedback in the event handler before starting a new load. */}
      <KBarProvider>
        <CommandPalette actions={actions} loading={loading} error={error} onRetry={() => { setError(null); setAttempt((value) => value + 1); }} />
        {/* The details route receives the same collection used by search.
            Unknown person IDs and unknown page URLs have separate fallbacks. */}
        <Routes>
          <Route index element={<Home dark={dark} onToggleTheme={() => setDark((value) => !value)} />} />
          <Route path="people/:id" element={<PersonDetails people={people} loading={loading} error={error} onPeopleChanged={refreshPeople} />} />
          {/* A new family has no person yet, so it shares the details layout
              without the shared people load the person route needs. */}
          <Route path="families/new" element={<NewFamily onPeopleChanged={refreshPeople} />} />
          <Route path="*" element={<main className="p-8"><h1>Page not found</h1><Link to="/">Go home</Link></main>} />
        </Routes>
      </KBarProvider>
    </Theme>
  );
}
