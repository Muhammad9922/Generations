import { useEffect, useMemo, useState } from "react";
import { Link, Route, Routes, useNavigate } from "react-router";
import PersonDetails from "./pages/PersonDetails";
import { Theme } from "@radix-ui/themes";
import { KBarProvider, type Action } from "kbar";
import { Home as HomeIcon, Moon } from "lucide-react";
import Home from "./pages/Home";
import CommandPalette from "./components/CommandPalette";
import { createPersonActions, GetAllUsers, type Person } from "./helpers/GetUsers";

/**
 * Owns shared people data and the session's theme preference. Keeping the Kbar
 * provider above Routes makes search available on both home and details pages.
 * PersonDetails only renders data; CommandPalette registers and displays actions.
 */
export default function App() {
  // BrowserRouter in the entry point supplies client-side navigation.
  const navigate = useNavigate();
  // One load state feeds both search feedback and the details page.
  const [people, setPeople] = useState<Person[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  // Incrementing attempt explicitly reruns the loader after a failed request.
  const [attempt, setAttempt] = useState(0);
  // Theme is intentionally in-memory: refreshing restores the light default.
  const [dark, setDark] = useState(false);

  useEffect(() => {
    // Ignore settled promises from an obsolete effect, including StrictMode's
    // development setup/cleanup cycle. This flag does not abort a network call.
    let cancelled = false;
    GetAllUsers().then((data) => {
      if (!cancelled) setPeople(data);
    }).catch(() => {
      if (!cancelled) setError("Could not load people.");
    }).finally(() => {
      if (!cancelled) setLoading(false);
    });
    return () => { cancelled = true; };
  }, [attempt]);

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
        <CommandPalette actions={actions} loading={loading} error={error} onRetry={() => { setLoading(true); setError(null); setAttempt((value) => value + 1); }} />
        {/* The details route receives the same collection used by search.
            Unknown person IDs and unknown page URLs have separate fallbacks. */}
        <Routes>
          <Route index element={<Home dark={dark} onToggleTheme={() => setDark((value) => !value)} />} />
          <Route path="people/:id" element={<PersonDetails people={people} loading={loading} error={error} />} />
          <Route path="*" element={<main className="p-8"><h1>Page not found</h1><Link to="/">Go home</Link></main>} />
        </Routes>
      </KBarProvider>
    </Theme>
  );
}
