import { useRef } from "react";
import * as Menubar from "@radix-ui/react-menubar";
import { Theme } from "@radix-ui/themes";
import { CalendarDays, ChevronDown, Plus, Settings2, Trash2 } from "lucide-react";

interface Props {
  label: string;
  canAddChild: boolean;
  onAddChild: (trigger: HTMLButtonElement | null) => void;
  onEditDates: (trigger: HTMLButtonElement | null) => void;
  onDelete: (trigger: HTMLButtonElement | null) => void;
}

/** Actions target the adjacent marriage; dialogs open after menu focus closes. */
export default function MarriageMenubar({ label, canAddChild, onAddChild, onEditDates, onDelete }: Props) {
  const actionsRef = useRef<HTMLButtonElement>(null);
  const pendingAction = useRef<((trigger: HTMLButtonElement | null) => void) | null>(null);
  return <Menubar.Root className="marriage-menubar" aria-label={`${label} controls`}>
    <Menubar.Menu>
      <Menubar.Trigger ref={actionsRef} className="marriage-menu-trigger" aria-label={`${label} actions`}>
        <Settings2 size={16} aria-hidden="true" />Actions<ChevronDown size={14} aria-hidden="true" />
      </Menubar.Trigger>
      <Menubar.Portal><Theme>
        <Menubar.Content className="marriage-menu-content" side="bottom" align="end" sideOffset={8} collisionPadding={12}
          onCloseAutoFocus={(event) => {
            const action = pendingAction.current;
            if (action) { event.preventDefault(); pendingAction.current = null; action(actionsRef.current); }
          }}>
          <Menubar.Label className="marriage-menu-label">{label}</Menubar.Label>
          <Menubar.Item className="marriage-menu-item" disabled={!canAddChild} onSelect={() => { pendingAction.current = onAddChild; }}>
            <Plus size={16} aria-hidden="true" />Add child…
          </Menubar.Item>
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
  </Menubar.Root>;
}
