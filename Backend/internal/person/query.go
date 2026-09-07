package person

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

func GetPerson(ctx context.Context, driver neo4j.Driver, id string) (*NewPerson, error) {
	const query = `
    MATCH (p:Person {id: $id})
    RETURN p.name AS name, p.id AS id, p.gender AS gender, p.date_of_birth AS date_of_birth, p.alive AS alive
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
		p.id = val
	}
	if val, ok := personMap["gender"].(string); ok {
		p.Gender = Gender(val) // Convert string -> Gender custom type
	}
	if val, ok := personMap["date_of_birth"].(string); ok {
		p.DateOfBirth = DateOfBirth(val) // Handles missing optional date safely
	}
	if val, ok := personMap["alive"].(bool); ok {
		p.Alive = val
	}

	return p, nil
}

/*

func GetParents() []NewPerson {

}

func GetChildren() []NewPerson {

}

func GetSpouse() {

}

func GetConnectionBetweenTwoPerson() {

}

*/
