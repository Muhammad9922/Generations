package person

import (
	"context"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

// Helper function to safely format Neo4j date types or string representations to dd-mm-yyyy
func formatDate(val any) (string, bool) {
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
		p.Gender = Gender(val)
	}

	// Process Date of Birth
	if dateStr, ok := formatDate(personMap["date_of_birth"]); ok {
		p.DateOfBirth = DateProper(dateStr)
		if !p.DateOfBirth.IsValid() {
			return nil, fmt.Errorf("Invalid Date Of Birth: %s", dateStr)
		}
	}

	// Process Date of Death
	if dateStr, ok := formatDate(personMap["date_of_death"]); ok {
		p.DateOfDeath = DateProper(dateStr)
		if !p.DateOfDeath.IsValid() {
			return nil, fmt.Errorf("Invalid Date Of Death: %s", dateStr)
		}
	}

	if val, ok := personMap["alive"].(bool); ok {
		p.Alive = val
	}

	return p, nil
}
