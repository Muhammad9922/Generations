# Manual frontend test checklist

For the person details page and its marriage model.

**The app now reads everything from the Go API, so this checklist needs that API running.**
`Backend/cmd/server/main.go` still answers `"Hi there!"` on every path — until the endpoints in
`api-contracts/README.md` exist, the app shows "Could not reach the API." and nothing else.

## Setup

```powershell
cd C:\Users\rehab\Documents\Muhammad-Projects\Generations\Frontend
node tests/api.test.mjs; node tests/family.test.mjs; node tests/people.test.mjs   # expect 11 + 7 + 4 pass
npm run dev
```

Run the Go API on `:8080` (Vite proxies `/api` to it, so no CORS handling is needed). Open the
printed URL (Vite default `http://localhost:5173`), open DevTools, and filter the console on
`api-contract` to see every request and response.

## Data

The IDs below (`id-1` … `id-9`) came from the built-in sample family, which now lives only in
`tests/sampleFamily.mjs`. Substitute IDs from your own database: the flows and the expected
behaviour are unchanged, but which person has two marriages, no parents or four children depends on
what you have stored. Where a check names a person, it also states the shape it needs — e.g. "a
person with two marriages and two children in each".

---

## 1. Home page and command palette

- [ ] Home shows "Welcome To Generations!" with four cards.
- [ ] "List Of Singles" and "New Family" are **disabled** and labelled "(coming soon)".
- [ ] "Search User" opens the palette; "Dark Mode" / "Light Mode" toggles the theme and the label flips.
- [ ] `Ctrl+K` / `⌘K` opens the palette from Home **and** from a person page.
- [ ] Palette shows sections `Navigation`, `Preferences`, `People`.
- [ ] Typing `Usman` shows both duplicates as "Usman Abdul son of Abdul Rahman" / "…son of Abdul Rauf".
- [ ] Typing `mahammad` (lowercase) still matches — search is case-insensitive and also matches father names and IDs.
- [ ] Typing `zzzz` shows "No matches. Try a name or father's name."
- [ ] `↑`/`↓` move the highlight, `Enter` runs the entry, `Esc` closes, the `Esc` button closes.
- [ ] `Go home` navigates to `/`; a person result navigates to `/people/<id>`.
- [ ] Footer reads "↑ ↓ navigate · Enter select · Esc close" and mentions nothing about sample, demo or mock data.

## 2. Opening a person (the payload)

- [ ] `/people/id-1` shows a "Loading person…" state, then the family view.
- [ ] Console has a collapsed `[api-contract] Request: GET /people/:id` group and a `Response` group with `Duration` and `found: true`.
- [ ] Response payload contains `Person`, `ParentsMarriage` (with `Spouse` + `Children` holding this person) and `Marriages`, and each person shows `id` / `name` / `dateOfBirth` / `dateOfDeath` / `gender` / `alive` rather than `undefined`.
- [ ] `/people/does-not-exist` shows "Person not found" (not a crash, not "Loading").
- [ ] `/nonsense` shows "Page not found" with a "Go home" link.
- [ ] "Back home" returns to `/` without a page reload.
- [ ] The notice under the person's name explains the marriage model and says nothing about mocks, samples or demo data.

## 3. Parents section

- [ ] On `id-1`: two parent cards, headed "Parents · the marriage they are a child of".
- [ ] On `id-2`: instead of cards, a dashed **Add parents** tile.
- [ ] Hovering a parent card does **not** highlight anything else (parents have no marriage colour in this view).
- [ ] Clicking a parent card opens "Edit person" with that parent's details, ID shown at the bottom.
- [ ] Parent card shows name, age, gender, "Born …", and "Died …" only when not alive.
- [ ] **Delete parents** appears only when parents exist; it does **not** appear on `id-2`.

## 4. Person and spouses

- [ ] On `id-1`: the focal card plus two marriage groups, labelled "Family 1 · Spouse" and "Family 2 · Spouse".
- [ ] Each marriage group has an **Actions** menu with "Add child…", "Change dates…", "Delink spouse…".
- [ ] Hovering a spouse card highlights the focal card and that marriage's children in the same family colour.
- [ ] Focusing a spouse card with `Tab` produces the same highlight (keyboard parity).
- [ ] Hovering a child highlights only that child's marriage group.
- [ ] **Change dates…** on Family 1 shows the start `12-06-2005` and end `15-10-2015` (as `2005-06-12` / `2015-10-15` in the date inputs).
- [ ] Clearing the end date and saving removes the end date from the card view; console logs `[api-contract] Request: PATCH /marriages/:id`.
- [ ] Entering an end date **before** the start date shows an inline error and keeps the dialog open.
- [ ] On `id-2`: no marriage groups, and the empty state says "Add a spouse before adding children."

## 5. Children section

- [ ] Heading shows the count and "Oldest first".
- [ ] `id-1` lists 4 children sorted oldest → youngest across both marriages (Maryam 2007, Yusuf 2014, Noor 2018, Adam 2021).
- [ ] Each child card has its own **Remove child** button underneath it (not inside the card).
- [ ] Clicking the child card opens "Edit person"; clicking **Remove child** opens the confirmation instead — the two controls do not overlap or swallow each other's clicks.
- [ ] On `id-3`: children section is empty with the hint "No children added yet. Use Actions beside a spouse to add a child."

## 6. Editing any person (focal, parent, spouse, child)

- [ ] Changing a **name** saves, and the heading / palette label update immediately.
- [ ] Clearing the name shows "Every person needs a name." and does not close the dialog.
- [ ] Typing `31-02-2000` as a date shows "Enter a valid calendar date."
- [ ] A future birth date is rejected; unchecking **Alive** reveals the death-date field.
- [ ] A death date before birth, or in the future, is rejected.
- [ ] The **ID** shown in the dialog never changes on save.
- [ ] Cancel discards the edit.

## 7. Add spouse

- [ ] **Add spouse** opens the dialog; the select lists only opposite-gender people who are not already a spouse.
- [ ] "Create new person" reveals name / birth / alive / death fields, with gender locked to the opposite of the focal person.
- [ ] Submitting without choosing anyone keeps the button disabled.
- [ ] A new spouse with no name is rejected ("The new spouse needs a name.").
- [ ] On success a new "Family N · Spouse" group appears; console shows `CreatePerson` then `CreateMarriage`.
- [ ] Marriage dates entered here appear on the new group.

## 8. Add child

- [ ] On `id-1`: **Actions → Add child…** opens the dialog; "Other parent" lists existing spouses first.
- [ ] **Add child** (bottom button) is enabled when the person has a spouse or an eligible other-gender person exists.
- [ ] The **Child** selector stays empty until an other parent is chosen.
- [ ] The **Child** selector groups candidates under "Younger than both spouses" and ends with a **Create new person** item.
- [ ] Only people younger than both spouses are listed. On `id-1` with Sara Ahmed (08-09-1983) as the other parent the candidates are exactly **Fatima Noor** (no birth date on record) and **Ibrahim Khalid** (09-07-1992): everyone born 14-03-1980 is the same age as this person and Layla Omar (1978) is older.
- [ ] Someone with no recorded birth date still appears — an unknown date cannot prove they are older (`id-7`, Fatima Noor).
- [ ] Neither spouse, anyone already married to this person, and anyone already a child of that marriage are listed.
- [ ] Picking an existing person shows "<name> · Age N · born DD-MM-YYYY" under the selector and the confirm button reads **Add child**.
- [ ] Submitting with no child chosen keeps the button disabled.
- [ ] Linking an existing person to an existing marriage adds them to that marriage's group; the console shows `AddChildren` with **no** `CreatePerson`.
- [ ] Choosing **Create new person** reveals name / gender / birth / alive / death, and the confirm button reads **Create child**.
- [ ] A new child with no name is rejected ("The child needs a name."); creating one on an existing marriage logs `CreatePerson` → `AddChildren`.
- [ ] Choosing "Other eligible people" (no existing spouse) shows the **Parents' marriage dates** fieldset; the new marriage appears with the child in it (`CreatePerson` → `CreateMarriage`).
- [ ] "Add a spouse for this child" → "Create new person" creates the child **and** their spouse (two `CreatePerson` + two `CreateMarriage`).
- [ ] After that flow the child's own new marriage does **not** appear in "Person & spouses" — that area only ever lists marriages *this* person is a spouse in.
- [ ] After any add-child flow, the new child appears in the Children list sorted by birth date.

## 9. Add parents

- [ ] Visible on `id-2`, `id-6`, `id-7` (people with no parents' marriage).
- [ ] Both Father and Mother must be chosen; **Add parents** stays disabled until both are selected.
- [ ] Creating a parent with a birth date **after** the child's birth is rejected.
- [ ] Picking the child as their own parent is not offered.
- [ ] Success: two parent cards appear and the Add-parents tile is replaced (console: `CreatePerson` ×1–2 → `CreateMarriage` with the child among its children).
- [ ] "Delete parents" now appears.

## 10. Destructive flows and focus

- [ ] **Delete parents** on `id-1` asks "Delete parents?" and quotes the person's name; confirming replaces the parent cards with the Add-parents tile.
  - [ ] After confirming, focus lands on the "Person & spouses" heading (the Delete-parents button no longer exists).
  - [ ] Cancel leaves the parents in place and returns focus to the Delete-parents button.
- [ ] **Delink spouse** on `id-1` Family 1 asks "Delink spouse?" and says the 2 children will be left without parents; confirming removes that group and those 2 children disappear from the Children list.
- [ ] **Delink spouse** `id-1` Family 2 (2 children) behaves the same; the marriage count in the heading drops.
- [ ] **Remove child** on any child asks "Remove child?" and, on confirm, removes only that child; the marriage group and both spouses stay; the count decrements.
- [ ] Console shows `RemoveChild` / `DeleteMarriage` for those actions.
- [ ] `Esc` closes every dialog and every menu without applying anything.

## 11. Resets and persistence

- [ ] Search **renames** survive navigating away and back to the person (until a full page refresh).
- [ ] Relationship edits (added/removed parents, spouses, children, changed dates) **reset** when you navigate away and return — by design, the view is rebuilt from fixtures.
- [ ] A full page refresh resets everything, including renames.
- [ ] Removing a child or delinking a spouse does **not** remove the person from search results.
- [ ] Deleting a marriage renumbers the remaining groups ("Family 2" becomes "Family 1").

## 12. Keyboard and assistive tech

- [ ] `Tab` reaches: Back home → parent cards → Add parents → focal card → spouse cards → Actions menu → Add spouse → Add child → child cards → Remove child buttons.
- [ ] Every card announces as "Edit <name>, <label>".
- [ ] Focus rings are visible on cards, add tiles, remove buttons and the Actions trigger.
- [ ] The Actions menu opens with `Enter`, moves with arrow keys, and closes with `Esc` returning focus to the trigger.
- [ ] Dialogs are named by their title and trap focus while open.
- [ ] Parents/Children headings announce their helper text ("the marriage they are a child of", the count).

## 13. Not testable / expected limitations

- [ ] **Loading and error states are reachable now**: stop the Go API and reload — the palette reports the failure and **Retry** re-runs the load; start the API again and Retry recovers. The same is true of a person page, which reports its own load failure.
- [ ] Newly created people do **not** appear in the palette or in other dialogs' selectors until a full page reload.
- [ ] The palette subtitle still shows the fixture `personFatherName`; adding or renaming a parent does not change it.
- [ ] There is no way to replace or edit an existing parents' marriage — only Delete parents, then Add parents.
- [ ] A marriage's spouses can never be edited — only its dates, or Delink spouse.

## 14. Fixed in this revision — confirm it stays fixed

**A child's own marriage must not leak into the parent's payload**

1. Open `/people/id-5` (has a parents' marriage, no spouse).
2. Click **Add child**.
3. Under "Other parent", pick someone from **Other eligible people** (not an existing spouse).
4. Under the new child, set "Add a spouse for this child" to **Create new person**, fill in the name.
5. Click **Create child**.

- [ ] Exactly one new marriage appears under "Person & spouses": the one with the chosen other parent.
- [ ] The child and the child's spouse are **not** listed as this person's spouses.
- [ ] The new child appears once in the Children list.

## 15. Opening any person as the primary person

Every card except the selected person's own carries an **Open** button.

- [ ] Parents: **Open** on a parent goes to `/people/<their id>`; that page shows the parent with "Family 1" (this person's parents' marriage), the other parent as their spouse, and this person in the Children list. Their Parents section shows the Add-parents tile.
- [ ] Spouses: **Open** on a spouse shows them with the shared marriage, its children, and no parents' marriage.
- [ ] Children: **Open** on a child shows them with the marriage that produced them as their parents' marriage — both parents listed — and no marriages of their own, so the Children section reads "Add a spouse before adding children."
- [ ] The selected person's own card has **no** Open button.
- [ ] The URL carries the person's ID, so the page can be reloaded or pasted into a new tab and still resolves (`/people/id-1-child-1`, `/people/id-1-spouse-1`, `/people/id-1-parent-1`).
- [ ] Navigating between relatives keeps the palette shortcut (`Ctrl+K`) working, and "Back home" still returns to `/`.
- [ ] A person opened as primary shows the same family as they do as a card on the page you came from.
- [ ] A marriage or other non-person ID in the URL (`/people/id-1-marriage-1`) shows "Person not found", not a crash or an empty family.
