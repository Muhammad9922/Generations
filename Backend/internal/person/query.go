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

// personFields is the projection every person query returns. Sharing it keeps
// the queries and scanPerson from drifting apart: a column added to one and not
// the other would silently read as a missing property.
const personFields = `p.name AS name, p.id AS id, p.gender AS gender, p.date_of_birth AS date_of_birth, p.alive AS alive, p.date_of_death AS date_of_death`

// scanPerson turns one query row into a person, checking the two keys the graph
// must carry and normalising both dates to DD-MM-YYYY.
//
// GetPerson and GetPersonList used to hold a copy of this each, which is how the
// two could disagree about a date that arrived as a string rather than a
// neo4j.Date.
func scanPerson(record *neo4j.Record) (NewPerson, error) {
	personMap := record.AsMap()

	for _, key := range []string{"name", "id"} {
		if personMap[key] == nil {
			return NewPerson{}, fmt.Errorf("required key missing: %s", key)
		}
	}

	var p NewPerson

	if val, ok := personMap["name"].(string); ok {
		p.PersonName = val
	}
	if val, ok := personMap["id"].(string); ok {
		p.Id = val
	}
	if val, ok := personMap["gender"].(string); ok {
		p.Gender = Gender(val)
	}

	birth, err := scanDate(personMap["date_of_birth"], "Date Of Birth")
	if err != nil {
		return NewPerson{}, err
	}
	p.DateOfBirth = birth

	death, err := scanDate(personMap["date_of_death"], "Date Of Death")
	if err != nil {
		return NewPerson{}, err
	}
	p.DateOfDeath = death

	if val, ok := personMap["alive"].(bool); ok {
		p.Alive = val
	}

	return p, nil
}

// scanDate accepts the DD-MM-YYYY the queries are written to return, plus the
// YYYY-MM-DD that older records store, and returns the canonical DD-MM-YYYY.
func scanDate(raw any, field string) (DateProper, error) {
	if raw == nil {
		return "", nil
	}

	// fmt.Sprint converts a neo4j.Date (or any Stringer) without a type switch.
	parts := strings.Split(fmt.Sprint(raw), "-")
	if len(parts) < 3 {
		return "", fmt.Errorf("Invalid %s: %v", field, raw)
	}

	// A year first means the stored value is YYYY-MM-DD.
	if len(parts[0]) > 2 {
		parts[0], parts[2] = parts[2], parts[0]
	}

	value := DateProper(strings.Join(parts, "-"))
	if !value.IsValid() {
		return "", fmt.Errorf("Invalid %s: %s", field, value)
	}

	return value, nil
}

func GetPerson(ctx context.Context, driver neo4j.Driver, id string) (*NewPerson, error) {
	const query = `
    MATCH (p:Person {id: $id})
    RETURN ` + personFields + `
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

	found, err := scanPerson(result.Records[0])
	if err != nil {
		return nil, err
	}

	return &found, nil
}

func GetPersonList(ctx context.Context, driver neo4j.Driver) ([]NewPerson, error) {
	const query = `
	MATCH (p:Person)
    RETURN ` + personFields + `
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

	return scanPeople(result.Records)
}

// GetSingleList returns everyone who is not a spouse in any marriage.
//
// A person is a spouse by holding a MARRIED edge to a Marriage node, so being
// single means having no such edge. Disbanding a marriage therefore makes its
// spouses single again without the person nodes changing, and a widowed spouse
// counts as single: this graph records that a marriage existed, not that it
// ended.
//
// The obvious `WHERE NOT (p)-[:MARRIED]->(:Marriage)` is not usable here: the
// store is Memgraph, which rejects a pattern as an atom expression ("Not yet
// implemented") and then retries until the driver's budget runs out. Counting an
// optional match asks the same question without one.
func GetSingleList(ctx context.Context, driver neo4j.Driver) ([]NewPerson, error) {
	const query = `
	MATCH (p:Person)
	OPTIONAL MATCH (p)-[:MARRIED]->(m:Marriage)
	WITH p, count(m) AS marriageCount
	WHERE marriageCount = 0
	RETURN ` + personFields + `
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

	return scanPeople(result.Records)
}

// scanPeople maps a whole result set, stopping at the first row that cannot be
// read rather than returning a list with a hole in it.
func scanPeople(records []*neo4j.Record) ([]NewPerson, error) {
	people := make([]NewPerson, 0, len(records))

	for _, record := range records {
		p, err := scanPerson(record)
		if err != nil {
			return nil, err
		}
		people = append(people, p)
	}

	return people, nil
}
