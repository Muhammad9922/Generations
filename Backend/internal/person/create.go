package person

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

type Gender string

const (
	Male   Gender = "Male"
	Female Gender = "Female"
)

type NewPerson struct {
	PersonName  string
	Gender      Gender
	DateOfBirth string
	ParentID    string
	Alive       bool
}

var dateRegix = regexp.MustCompile(`^(0[1-9]|[12][0-9]|3[01])-(0[1-9]|1[0-2])-\d{4}$`)

func CreateNewPerson(ctx context.Context, driver neo4j.Driver, params NewPerson) (string, error) {
	// Name Is Mandatory
	if params.PersonName == "" {
		return "", errors.New("The Person's Name Must Be Given")
	}

	if params.Gender == "" {
		return "", errors.New("A Gender Must Be Supplied For New Person")
	}

	if params.DateOfBirth != "" {
		correct_date := dateRegix.MatchString(params.DateOfBirth)
		if correct_date == false {
			return "", errors.New("The Date Of Birth Must Be In DD-MM-YYYY Format")
		}
	}

	if params.ParentID != "" {
		exists, err := CheckPersonExistence(ctx, driver, PersonQuery{
			ID: params.ParentID,
		})
		if err != nil {
			return "", fmt.Errorf("failed to verify parent existence: %w", err)
		}
		if !exists {
			return "", fmt.Errorf("parent with ID %q does not exist", params.ParentID)
		}
	}

	if params.ParentID != "" {
		neo4j.ExecuteQuery(
			ctx,
			driver,
			`MERGE (p:Person {name: $name, gender: $gender, DateOfBirth: $dob, ParentID: $parentid})`,
			map[string]any{
				"name":     params.PersonName,
				"gender":   params.Gender,
				"dob":      params.DateOfBirth,
				"parentid": params.ParentID,
			},
			neo4j.EagerResultTransformer,
		)
	} else {
		neo4j.ExecuteQuery(
			ctx,
			driver,
			`MERGE (p:Person {name: $name, gender: $gender, DateOfBirth: $dob})`,
			map[string]any{
				"name":   params.PersonName,
				"gender": params.Gender,
				"dob":    params.DateOfBirth,
			},
			neo4j.EagerResultTransformer,
		)
	}

	return params.PersonName, nil
}
