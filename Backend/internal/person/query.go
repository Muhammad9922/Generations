package person

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

// Helper function to safely format Neo4j date types or string representations to dd-mm-yyyy
func FormatDate(val any) (string, bool) {
	if val == nil {
		return "", false
	}
	switch v := val.(type) {
	case neo4j.Date:
		return v.Time().Format("02-01-2006"), true
	case time.Time:
		return v.Format("02-01-2006"), true
	case string:
		// Handles cases where date is already a string in YYYY-MM-DD
		if t, err := time.Parse("2006-01-02", v); err == nil {
			return t.Format("02-01-2006"), true
		}
		return v, true
	default:
		return "", false
	}
}

// ymdRegex captures YYYY-MM-DD and dmyRegex captures DD-MM-YYYY. The two forms
// are unambiguous, so a value can only match one of them.
var ymdRegex = regexp.MustCompile(`^(\d{4})-(\d{2})-(\d{2})$`)
var dmyRegex = regexp.MustCompile(`^(\d{2})-(\d{2})-(\d{4})$`)

// ParseDateProper accepts a date written as either DD-MM-YYYY or YYYY-MM-DD and
// returns it in the canonical DD-MM-YYYY form used throughout this package.
//
// Unlike NormalizeDate it fully validates the value: the components must be in
// range (no month 13) and the day must exist in that month (31-02-2023 is
// rejected). Any other format is rejected.
func ParseDateProper(input string) (DateProper, error) {
	normalized, err := NormalizeDate(input)
	if err != nil {
		return "", err
	}

	// normalized is guaranteed to be YYYY-MM-DD by NormalizeDate.
	parts := ymdRegex.FindStringSubmatch(normalized)
	if parts == nil {
		return "", fmt.Errorf("invalid date %q: must be YYYY-MM-DD or DD-MM-YYYY", input)
	}

	candidate := DateProper(fmt.Sprintf("%s-%s-%s", parts[3], parts[2], parts[1]))

	if !candidate.IsValid() {
		return "", fmt.Errorf("invalid date %q: not a real date in DD-MM-YYYY or YYYY-MM-DD format", input)
	}

	// IsValid is regex-only, so it still accepts 31-02-2023; ParseTime rejects
	// days that do not exist in the given month.
	if _, ok := candidate.ParseTime(); !ok {
		return "", fmt.Errorf("invalid date %q: that calendar day does not exist", input)
	}

	return candidate, nil
}

// NormalizeDate takes a date in either YYYY-MM-DD or DD-MM-YYYY and always
// returns it as YYYY-MM-DD.
//
// NOTE: the result is ISO formatted, which the DateProper helpers (IsValid,
// ParseTime, GetNeoDate) do NOT accept — they expect DD-MM-YYYY. Use
// ParseDateProper when you need a validated DateProper.
func NormalizeDate(dateStr string) (string, error) {
	// Check if it is already in YYYY-MM-DD format
	if ymdRegex.MatchString(dateStr) {
		return dateStr, nil
	}

	// Check if it matches DD-MM-YYYY and extract the parts.
	// FindStringSubmatch returns a slice of the captured groups inside the parentheses ()
	matches := dmyRegex.FindStringSubmatch(dateStr)

	if matches != nil {
		// matches[0] is the full matched string ("19-09-2026")
		// matches[1] is the DD part ("19")
		// matches[2] is the MM part ("09")
		// matches[3] is the YYYY part ("2026")

		// Reformat and return as YYYY-MM-DD
		return fmt.Sprintf("%s-%s-%s", matches[3], matches[2], matches[1]), nil
	}

	// Neither regex matched
	return "", fmt.Errorf("invalid date format %q: must be YYYY-MM-DD or DD-MM-YYYY", dateStr)
}

func GetPerson(ctx context.Context, driver neo4j.Driver, id string) (*NewPerson, error) {
	const query = `
    MATCH (p:Person {id: $id})
    RETURN p.name AS name, p.id AS id, p.gender AS gender, p.date_of_birth AS date_of_birth, p.alive AS alive, p.date_of_death as date_of_death
    `
	params := map[string]any{
		"id": id,
	}

	result, err := neo4j.ExecuteQuery(
		ctx,
		driver,
		query,
		params,
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return nil, err
	}

	if len(result.Records) == 0 {
		return nil, fmt.Errorf("%w: person not found with id: %s", ErrNotFound, id)
	}

	personMap := result.Records[0].AsMap()

	requiredKeys := []string{"name", "id"}
	for _, key := range requiredKeys {
		if personMap[key] == nil {
			return nil, fmt.Errorf("required key missing: %s", key)
		}
	}

	p := &NewPerson{}

	if val, ok := personMap["name"].(string); ok {
		p.PersonName = val
	}
	if val, ok := personMap["id"].(string); ok {
		p.Id = val
	}
	if val, ok := personMap["gender"].(string); ok {
		p.Gender = Gender(val) // Convert string -> Gender custom type
	}

	if dobRaw, ok := personMap["date_of_birth"]; ok && dobRaw != nil {
		dobStr := fmt.Sprint(dobRaw) // Converts neo4j.Date / Stringer to string safely
		dobStrSplit := strings.Split(dobStr, "-")
		var dobStrFixed string
		if len(dobStrSplit) < 3 {
			return nil, fmt.Errorf("Invalid Date Of Birth: %v", dobStr)
		}
		if len(dobStrSplit[0]) > 2 {
			dobStrFixed = fmt.Sprintf("%v-%v-%v", dobStrSplit[2], dobStrSplit[1], dobStrSplit[0])
		} else {
			dobStrFixed = fmt.Sprintf("%v-%v-%v", dobStrSplit[0], dobStrSplit[1], dobStrSplit[2])
		}

		properDob := DateProper(dobStrFixed)
		if !properDob.IsValid() {
			return nil, fmt.Errorf("Invalid Date Of Birth: %s", properDob)
		}
		p.DateOfBirth = properDob
	}

	if dodRaw, ok := personMap["date_of_death"]; ok && dodRaw != nil {
		dodStr := fmt.Sprint(dodRaw) // Converts neo4j.Date / Stringer to string safely
		dodStrSplit := strings.Split(dodStr, "-")
		var dodStrFixed string

		if len(dodStrSplit) < 3 {
			return nil, fmt.Errorf("Invalid Date Of Death: %v", dodStr)
		}

		if len(dodStrSplit[0]) > 2 {
			dodStrFixed = fmt.Sprintf("%v-%v-%v", dodStrSplit[2], dodStrSplit[1], dodStrSplit[0])
		} else {
			dodStrFixed = fmt.Sprintf("%v-%v-%v", dodStrSplit[0], dodStrSplit[1], dodStrSplit[2])
		}
		properDod := DateProper(dodStrFixed)
		if !properDod.IsValid() {
			return nil, fmt.Errorf("Invalid Date Of Death: %s", properDod)
		}
		p.DateOfDeath = properDod
	}

	if val, ok := personMap["alive"].(bool); ok {
		p.Alive = val
	}

	return p, nil
}

func GetPersonList(ctx context.Context, driver neo4j.Driver) ([]NewPerson, error) {
	const query = `
	MATCH (p:Person)
    RETURN p.name AS name, p.id AS id, p.gender AS gender, p.date_of_birth AS date_of_birth, p.alive AS alive, p.date_of_death as date_of_death
	`

	result, err := neo4j.ExecuteQuery(
		ctx,
		driver,
		query,
		map[string]any{},
		neo4j.EagerResultTransformer,
	)

	if err != nil {
		return nil, err
	}

	var people []NewPerson

	records := (*result).Records

	for _, record := range records {
		personMap := record.AsMap()

		requiredKeys := []string{"name", "id"}
		for _, key := range requiredKeys {
			if personMap[key] == nil {
				return nil, fmt.Errorf("required key missing: %s", key)
			}
		}

		p := &NewPerson{}

		if val, ok := personMap["name"].(string); ok {
			p.PersonName = val
		}
		if val, ok := personMap["id"].(string); ok {
			p.Id = val
		}
		if val, ok := personMap["gender"].(string); ok {
			p.Gender = Gender(val) // Convert string -> Gender custom type
		}

		if dobRaw, ok := personMap["date_of_birth"]; ok && dobRaw != nil {
			dobStr := fmt.Sprint(dobRaw) // Converts neo4j.Date / Stringer to string safely
			dobStrSplit := strings.Split(dobStr, "-")
			var dobStrFixed string
			if len(dobStrSplit) < 3 {
				return nil, fmt.Errorf("Invalid Date Of Birth: %v", dobStr)
			}
			if len(dobStrSplit[0]) > 2 {
				dobStrFixed = fmt.Sprintf("%v-%v-%v", dobStrSplit[2], dobStrSplit[1], dobStrSplit[0])
			} else {
				dobStrFixed = fmt.Sprintf("%v-%v-%v", dobStrSplit[0], dobStrSplit[1], dobStrSplit[2])
			}

			properDob := DateProper(dobStrFixed)
			if !properDob.IsValid() {
				return nil, fmt.Errorf("Invalid Date Of Birth: %s", properDob)
			}
			p.DateOfBirth = properDob
		}

		if dodRaw, ok := personMap["date_of_death"]; ok && dodRaw != nil {
			dodStr := fmt.Sprint(dodRaw) // Converts neo4j.Date / Stringer to string safely
			dodStrSplit := strings.Split(dodStr, "-")
			var dodStrFixed string

			if len(dodStrSplit) < 3 {
				return nil, fmt.Errorf("Invalid Date Of Death: %v", dodStr)
			}

			if len(dodStrSplit[0]) > 2 {
				dodStrFixed = fmt.Sprintf("%v-%v-%v", dodStrSplit[2], dodStrSplit[1], dodStrSplit[0])
			} else {
				dodStrFixed = fmt.Sprintf("%v-%v-%v", dodStrSplit[0], dodStrSplit[1], dodStrSplit[2])
			}
			properDod := DateProper(dodStrFixed)
			if !properDod.IsValid() {
				return nil, fmt.Errorf("Invalid Date Of Death: %s", properDod)
			}
			p.DateOfDeath = properDod
		}

		if val, ok := personMap["alive"].(bool); ok {
			p.Alive = val
		}

		people = append(people, *p)
	}

	return people, nil
}
