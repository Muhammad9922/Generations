import type { CSSProperties } from "react";
import { UserRound, Pencil, ArrowUpRight, X } from "lucide-react";
import { ageLabel, familyColor, type User } from "../helpers/personModel.ts";

interface Props {
  user: User;
  label: string;
  familyId?: string;
  highlighted: boolean;
  onHover: (id: string | null) => void;
  onFocus: (id: string | null) => void;
  onEdit: (user: User) => void;
  /** Opens this person as the primary person on their own page. */
  onOpen?: (user: User) => void;
  /** Removes this person from the marriage they are shown under; the trigger restores focus. */
  onRemoveChild?: (user: User, trigger: HTMLButtonElement) => void;
}

/**
 * Buttons support the same linked highlight with pointer and keyboard focus.
 * Actions sit beside the card rather than inside it, so the card itself stays a
 * single accessible control and each action is reachable by keyboard.
 */
export default function FamilyPersonCard({ user, label, familyId, highlighted, onHover, onFocus, onEdit, onOpen, onRemoveChild }: Props) {
  return (
    <div className="family-card">
      <button type="button" className="family-person" data-highlighted={highlighted}
        style={familyId ? { "--family-color": familyColor(familyId) } as CSSProperties : undefined}
        onPointerEnter={() => onHover(familyId ?? null)} onPointerLeave={() => onHover(null)}
        onFocus={() => onFocus(familyId ?? null)} onBlur={() => onFocus(null)}
        onClick={() => onEdit(user)} aria-label={`Edit ${user.name}, ${label}`}>
        <span className="family-card-top"><span className="family-avatar"><UserRound size={22} /></span><Pencil size={14} /></span>
        <span className="family-eyebrow">{label}</span>
        <strong>{user.name}</strong>
        <span>{ageLabel(user)} · {user.gender}</span>
        <span>Born {user.dateOfBirth || "unknown"}</span>
        {!user.alive && <span>Died {user.dateOfDeath || "unknown"}</span>}
        <small>ID: {user.id}</small>
      </button>
      {(onOpen || onRemoveChild) && <div className="family-card-actions">
        {onOpen && <button type="button" className="family-open" onClick={() => onOpen(user)} aria-label={`Open ${user.name} as the primary person`}>
          <ArrowUpRight size={14} aria-hidden="true" />Open
        </button>}
        {onRemoveChild && <button type="button" className="family-remove" onClick={(event) => onRemoveChild(user, event.currentTarget)} aria-label={`Remove ${user.name} from ${label}`}>
          <X size={14} aria-hidden="true" />Remove child
        </button>}
      </div>}
    </div>
  );
}
