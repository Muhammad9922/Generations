package person

import (
	"context"
	"fmt"
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
		return nil, fmt.Errorf("person not found with id: %s", id)
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
