import { useEffect, useId, useMemo, useRef, useState } from "react";
import * as Popover from "@radix-ui/react-popover";
import { Theme } from "@radix-ui/themes";
import { ChevronDown, Search } from "lucide-react";
import type { User } from "../helpers/personModel.ts";

/** A labelled block of people. A group without a label renders as a flat run. */
export interface PersonOptionGroup {
  label?: string;
  people: User[];
}

interface Props {
  value: string;
  onChange: (value: string) => void;
  /** Shown while nothing is chosen; it is a placeholder, not a choice. */
  placeholder: string;
  groups: PersonOptionGroup[];
  /** Sentinel choices such as "Create new person", always listed last. */
  extras?: { value: string; label: string }[];
  disabled?: boolean;
  /** Accessible name; the visible field label is not enough once the popup opens. */
  label: string;
}

/** Headers and options share one row height so the window is simple arithmetic. */
type Row = { kind: "header"; label: string } | { kind: "option"; value: string; label: string };

const ROW_HEIGHT = 34;
const PADDING = 6;
const OVERSCAN = 4;
/**
 * How much of the list we assume is visible when deciding which rows to render.
 * The stylesheet caps the real height at or below this, so the window always
 * covers the viewport and no blank rows can appear while scrolling.
 */
const WINDOW_HEIGHT = 264;

/**
 * A searchable person picker.
 *
 * Radix's `Select` mounts every option as a component, so opening a list from a
 * 1302-person database blocked the main thread for ~550ms and made scrolling
 * janky. This keeps the same look but renders only the visible rows and lets the
 * list be typed into, so both opening and scrolling touch a handful of nodes.
 *
 * The popup rides Radix's `Popover` primitive (the themed `Popover.Anchor`
 * discards its children), which portals it clear of the dialog's clipping,
 * follows the trigger when space runs out, and — as the topmost dismissable
 * layer — takes the Escape key for itself instead of letting it close the whole
 * dialog. The primitive rather than the themed component keeps the trigger a
 * real `combobox`, which the themed Trigger would overwrite with
 * `aria-haspopup="dialog"`.
 */
export default function PersonPicker({ value, onChange, placeholder, groups, extras = [], disabled, label }: Props) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [highlighted, setHighlighted] = useState(-1);
  const [scrollTop, setScrollTop] = useState(0);
  const trigger = useRef<HTMLButtonElement>(null);
  const search = useRef<HTMLInputElement>(null);
  const list = useRef<HTMLDivElement>(null);
  const listId = useId();
  const optionId = (index: number) => `${listId}-option-${index}`;

  const selectedLabel = useMemo(() => {
    for (const group of groups) {
      const found = group.people.find((person) => person.id === value);
      if (found) return found.name;
    }
    return extras.find((extra) => extra.value === value)?.label ?? "";
  }, [groups, extras, value]);

  /** Headers and people that match the query, then the always-available sentinels. */
  const rows = useMemo<Row[]>(() => {
    const needle = query.trim().toLowerCase();
    const matches = (person: User) => !needle || person.name.toLowerCase().includes(needle) || person.id.toLowerCase().includes(needle);
    const built: Row[] = [];
    for (const group of groups) {
      const people = group.people.filter(matches);
      if (!people.length) continue;
      if (group.label) built.push({ kind: "header", label: group.label });
      for (const person of people) built.push({ kind: "option", value: person.id, label: person.name });
    }
    for (const extra of extras) built.push({ kind: "option", value: extra.value, label: extra.label });
    return built;
  }, [groups, extras, query]);

  const selectable = useMemo(() => rows.map((row, index) => row.kind === "option" ? index : -1).filter((index) => index >= 0), [rows]);
  // Derived rather than stored: a narrowed list can leave the remembered row on
  // a header or past the end, so fall back to the first real option.
  const activeIndex = selectable.includes(highlighted) ? highlighted : (selectable[0] ?? -1);

  // Keep the highlighted row inside the scrolled window during keyboard moves.
  useEffect(() => {
    const element = list.current;
    if (!open || !element || activeIndex < 0) return;
    const top = activeIndex * ROW_HEIGHT + PADDING;
    const bottom = top + ROW_HEIGHT;
    if (top < element.scrollTop) element.scrollTop = top;
    else if (bottom > element.scrollTop + element.clientHeight) element.scrollTop = bottom - element.clientHeight;
  }, [activeIndex, open]);

  function openPicker() {
    setQuery("");
    setScrollTop(0);
    setHighlighted(-1);
    setOpen(true);
  }

  function choose(next: string) {
    onChange(next);
    setOpen(false);
  }

  function move(step: number) {
    if (!selectable.length) return;
    const current = selectable.indexOf(activeIndex);
    // Wrap around, which is what a native listbox does at either end.
    setHighlighted(selectable[current === -1 ? 0 : (current + step + selectable.length) % selectable.length]);
  }

  function onSearchKeyDown(event: React.KeyboardEvent<HTMLInputElement>) {
    if (event.key === "ArrowDown") { event.preventDefault(); move(1); }
    else if (event.key === "ArrowUp") { event.preventDefault(); move(-1); }
    else if (event.key === "Home") { event.preventDefault(); setHighlighted(selectable[0] ?? -1); }
    else if (event.key === "End") { event.preventDefault(); setHighlighted(selectable[selectable.length - 1] ?? -1); }
    else if (event.key === "Enter") {
      // Always prevent: this input sits in the dialog's form, and Enter here
      // means "take the highlighted person", never "submit".
      event.preventDefault();
      const row = rows[activeIndex];
      if (row?.kind === "option") choose(row.value);
    } else if (event.key === "Tab") setOpen(false);
    // Escape is left to Radix: as the topmost layer it closes this popup only.
  }

  const start = Math.max(0, Math.floor(scrollTop / ROW_HEIGHT) - OVERSCAN);
  const end = Math.min(rows.length, Math.ceil((scrollTop + WINDOW_HEIGHT) / ROW_HEIGHT) + OVERSCAN);
  const windowHeight = rows.length * ROW_HEIGHT + PADDING * 2;

  return (
    <Popover.Root open={open} onOpenChange={(next) => (next ? openPicker() : setOpen(false))}>
      {/* Anchor rather than Trigger: the button stays a plain combobox and
          keeps its own ARIA, while the content positions itself against it. */}
      <Popover.Anchor asChild>
        <button type="button" ref={trigger} role="combobox" className="person-picker-trigger"
          aria-expanded={open} aria-haspopup="listbox" aria-controls={open ? listId : undefined}
          aria-label={label} disabled={disabled} data-placeholder={selectedLabel ? "false" : "true"}
          onClick={() => (open ? setOpen(false) : openPicker())}
          onKeyDown={(event) => {
            if (!open && (event.key === "ArrowDown" || event.key === "Enter" || event.key === " ")) { event.preventDefault(); openPicker(); }
          }}>
          <span>{selectedLabel || placeholder}</span>
          <ChevronDown size={16} aria-hidden="true" />
        </button>
      </Popover.Anchor>
      <Popover.Portal>
        {/* Theme re-establishes the Radix variables inside the portal. */}
        <Theme asChild>
          <Popover.Content className="person-picker-popup" side="bottom" align="start" sideOffset={6} collisionPadding={12}
            onOpenAutoFocus={(event) => { event.preventDefault(); search.current?.focus(); }}
            onCloseAutoFocus={(event) => { event.preventDefault(); trigger.current?.focus(); }}>
            <div className="person-picker-search">
              <Search size={15} aria-hidden="true" />
              <input ref={search} className="person-picker-input" value={query} autoComplete="off" spellCheck="false"
                placeholder="Type a name…" aria-label={`Search ${label.toLowerCase()}`}
                role="combobox" aria-expanded="true" aria-controls={listId} aria-autocomplete="list"
                aria-activedescendant={activeIndex >= 0 ? optionId(activeIndex) : undefined}
                onChange={(event) => {
                  setQuery(event.target.value);
                  // A new filter starts at the top, so the group heading and the
                  // best match are visible instead of the old scroll offset.
                  setScrollTop(0);
                  if (list.current) list.current.scrollTop = 0;
                }} onKeyDown={onSearchKeyDown} />
            </div>
            <div className="person-picker-list" ref={list} id={listId} role="listbox" aria-label={label}
              onScroll={(event) => setScrollTop(event.currentTarget.scrollTop)}>
              {rows.length === 0 ? <p className="person-picker-empty">No matches.</p> : (
                <div className="person-picker-window" style={{ height: windowHeight }}>
                  {rows.slice(start, end).map((row, offset) => {
                    const index = start + offset;
                    const style = { height: ROW_HEIGHT, transform: `translateY(${index * ROW_HEIGHT + PADDING}px)` };
                    return row.kind === "header"
                      ? <div key={`${row.label}-${index}`} className="person-picker-group" style={style} aria-hidden="true">{row.label}</div>
                      : <div key={row.value} id={optionId(index)} role="option" aria-selected={index === activeIndex}
                          className="person-picker-option" data-active={index === activeIndex} style={style}
                          onPointerEnter={() => setHighlighted(index)}
                          // Keep focus in the search box so typing can continue after a click.
                          onPointerDown={(event) => event.preventDefault()}
                          onClick={() => choose(row.value)}>{row.label}</div>;
                  })}
                </div>
              )}
            </div>
          </Popover.Content>
        </Theme>
      </Popover.Portal>
    </Popover.Root>
  );
}
