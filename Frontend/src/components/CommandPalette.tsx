import {
  KBarAnimator, KBarPortal, KBarPositioner, KBarResults, KBarSearch,
  useKBar, useMatches, useRegisterActions, type Action,
} from "kbar";
import { Search, User } from "lucide-react";
import { Theme } from "@radix-ui/themes";

/** App owns data and commands; the palette owns their search presentation. */
interface CommandPaletteProps {
  /** Memoized navigation, preference and person actions to register with Kbar. */
  actions: Action[];
  /** People may still be loading while navigation commands are already usable. */
  loading: boolean;
  /** Safe load-error message, independent of whether commands match the query. */
  error: string | null;
  /** Asks App to reset feedback and start another data-loading attempt. */
  onRetry: () => void;
}

/** Global command UI; must be rendered inside the app's KBarProvider. */
export default function CommandPalette({ actions, loading, error, onRetry }: CommandPaletteProps) {
  // Register asynchronous people and updated theme commands after initial mount.
  // The hook unregisters the previous list on change and cleans up on unmount.
  useRegisterActions(actions, [actions]);
  // Kbar performs matching and grouping; entries are section strings or actions.
  const { results } = useMatches();
  // Share the provider's controller instead of introducing a second open state.
  const { query } = useKBar();

  return (
    // Portal escapes page layout/overflow. Theme re-establishes Radix CSS
    // variables in the portaled DOM while inheriting the app's theme context.
    <KBarPortal>
      <Theme>
      {/* Kbar positions the overlay and animates its open/close transitions. */}
      <KBarPositioner className="command-backdrop">
        <KBarAnimator className="command-palette">
          <div className="command-search">
            <Search size={20} aria-hidden="true" />
            {/* Kbar controls the input, focus and combobox ARIA attributes.
                Its custom defaultPlaceholder prop supplies the root prompt. */}
            <KBarSearch defaultPlaceholder="Search people or commands…" aria-label="Search people or commands" />
            <button type="button" onClick={() => query.toggle()} aria-label="Close search">Esc</button>
          </div>
          {/* Announce request feedback without hiding usable static commands.
              Only show no-match feedback after loading finishes successfully. */}
          {loading && <p className="command-message" role="status">Loading people…</p>}
          {error && <div className="command-message" role="alert">{error} <button type="button" onClick={onRetry}>Retry</button></div>}
          {!loading && !error && results.length === 0 && (
            <p className="command-message" role="status">No matches. Try a name or father’s name.</p>
          )}
          {/* KBarResults owns virtualization, active selection and execution.
              Render section headers separately; data-active styles the current
              action for mouse and keyboard users without duplicating handlers.
              Actions may provide icons/shortcuts; people use the fallback icon. */}
          <KBarResults
            items={results}
            maxHeight={360}
            onRender={({ item, active }) => typeof item === "string" ? (
              <div className="command-section">{item}</div>
            ) : (
              <div className="command-result" data-active={active}>
                {item.icon ?? <User size={20} aria-hidden="true" />}
                <div className="command-label">
                  <div>{item.name}</div>
                  {item.subtitle && <small>{item.subtitle}</small>}
                </div>
                {item.shortcut?.map((key) => <kbd key={key}>{key}</kbd>)}
              </div>
            )}
          />
          <footer className="command-footer">↑ ↓ navigate · Enter select · Esc close</footer>
        </KBarAnimator>
      </KBarPositioner>
      </Theme>
    </KBarPortal>
  );
}
