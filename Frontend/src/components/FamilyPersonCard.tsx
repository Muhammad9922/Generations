import type { CSSProperties } from "react";
import { UserRound, Pencil } from "lucide-react";
import { ageLabel, familyColor, type User } from "../helpers/GetPersonDetails";

interface Props {
  user: User;
  label: string;
  familyId?: string;
  highlighted: boolean;
  onHover: (id: string | null) => void;
  onFocus: (id: string | null) => void;
  onEdit: (user: User) => void;
}

/** Buttons support the same linked highlight with pointer and keyboard focus. */
export default function FamilyPersonCard({ user, label, familyId, highlighted, onHover, onFocus, onEdit }: Props) {
  return (
    <button type="button" className="family-person" data-highlighted={highlighted}
      style={familyId ? { "--family-color": familyColor(familyId) } as CSSProperties : undefined}
      onPointerEnter={() => onHover(familyId ?? null)} onPointerLeave={() => onHover(null)}
      onFocus={() => onFocus(familyId ?? null)} onBlur={() => onFocus(null)}
      onClick={() => onEdit(user)} aria-label={`Edit ${user.Name}, ${label}`}>
      <span className="family-card-top"><span className="family-avatar"><UserRound size={22} /></span><Pencil size={14} /></span>
      <span className="family-eyebrow">{label}</span>
      <strong>{user.Name}</strong>
      <span>{ageLabel(user)} · {user.Gender}</span>
      <span>Born {user.DateOfBirth || "unknown"}</span>
      {!user.Alive && <span>Died {user.DeateOfDeath || "unknown"}</span>}
      <small>ID: {user.Id}</small>
    </button>
  );
}
