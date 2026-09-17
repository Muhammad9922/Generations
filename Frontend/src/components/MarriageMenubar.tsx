import { useRef, type RefObject } from "react";
import * as Menubar from "@radix-ui/react-menubar";
import { Theme } from "@radix-ui/themes";
import { CalendarDays, Check, ChevronDown, Heart, Settings2, Trash2 } from "lucide-react";
import { otherSpouses, type FamilyCertificate } from "../helpers/GetPersonDetails";

interface Props {
  families: FamilyCertificate[];
  selected?: FamilyCertificate;
  personId: string;
  onSelect: (id: string) => void;
  onEditDates: () => void;
  onDelete: () => void;
  actionsRef: RefObject<HTMLButtonElement | null>;
}

/** Separate menu selection from hover highlighting. Defer dialogs until menu focus closes. */
export default function MarriageMenubar({ families, selected, personId, onSelect, onEditDates, onDelete, actionsRef }: Props) {
  const pendingAction = useRef<(() => void) | null>(null);
  const name = (family: FamilyCertificate) => `Family ${families.findIndex((item) => item.Id === family.Id) + 1} · ${otherSpouses(family, personId).map((user) => user.Name).join(", ")}`;
  return <footer className="marriage-toolbar" tabIndex={-1} aria-label="Marriage controls">
    <Menubar.Root className="marriage-menubar" aria-label="Marriage controls" loop>
      <Menubar.Menu>
        <Menubar.Trigger className="marriage-menu-trigger" disabled={!families.length}>
          <Heart size={16} aria-hidden="true" />Marriage<ChevronDown size={14} aria-hidden="true" />
        </Menubar.Trigger>
        <Menubar.Portal><Theme>
          <Menubar.Content className="marriage-menu-content" side="top" align="end" sideOffset={8} collisionPadding={12}>
            <Menubar.Label className="marriage-menu-label">Choose a marriage</Menubar.Label>
            <Menubar.RadioGroup value={selected?.Id ?? ""} onValueChange={onSelect}>
              {families.map((family) => <Menubar.RadioItem className="marriage-menu-item" key={family.Id} value={family.Id} textValue={name(family)}>
                <span className="marriage-menu-check"><Menubar.ItemIndicator><Check size={16} aria-hidden="true" /></Menubar.ItemIndicator></span>
                {name(family)}
              </Menubar.RadioItem>)}
            </Menubar.RadioGroup>
          </Menubar.Content>
        </Theme></Menubar.Portal>
      </Menubar.Menu>
      <Menubar.Menu>
        <Menubar.Trigger ref={actionsRef} className="marriage-menu-trigger" disabled={!selected}>
          <Settings2 size={16} aria-hidden="true" />Actions<ChevronDown size={14} aria-hidden="true" />
        </Menubar.Trigger>
        <Menubar.Portal><Theme>
          <Menubar.Content className="marriage-menu-content" side="top" align="end" sideOffset={8} collisionPadding={12}
            onCloseAutoFocus={(event) => {
              const action = pendingAction.current;
              if (action) { event.preventDefault(); pendingAction.current = null; action(); }
            }}>
            <Menubar.Label className="marriage-menu-label">{selected ? name(selected) : "Marriage actions"}</Menubar.Label>
            <Menubar.Item className="marriage-menu-item" onSelect={() => { pendingAction.current = onEditDates; }}>
              <CalendarDays size={16} aria-hidden="true" />Change dates…
            </Menubar.Item>
            <Menubar.Separator className="marriage-menu-separator" />
            <Menubar.Item className="marriage-menu-item family-danger" onSelect={() => { pendingAction.current = onDelete; }}>
              <Trash2 size={16} aria-hidden="true" />Delete marriage…
            </Menubar.Item>
          </Menubar.Content>
        </Theme></Menubar.Portal>
      </Menubar.Menu>
    </Menubar.Root>
  </footer>;
}
