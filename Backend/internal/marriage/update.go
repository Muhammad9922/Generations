package marriage

import (
	"context"
	"errors"
	"fmt"

	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

type MarriageUpdate struct {
	DateStart person.DateProper
	DateEnd   person.DateProper
}

func UpdateMarriage(ctx context.Context, driver neo4j.Driver, id string, update MarriageUpdate) (string, error) {

	existingRecord, err := GetMarriage(ctx, driver, id)

	if err != nil {
		return "", err
	}

	if existingRecord == nil {
		return "", errors.New("No Existing Record Found")
	}

	if len(*existingRecord) != 1 {
		return "", errors.New("Unable To Properly Query Existing Marriage")
	}

	existingMarriage := (*existingRecord)[0]

	if update.DateStart == "" {
		update.DateStart = existingMarriage.Start
	}

	if update.DateEnd == "" {
		update.DateEnd = existingMarriage.End
	}

	if update.DateStart != "" && !update.DateStart.IsValid() {
		return "", fmt.Errorf("New DateStart Is Invalid %v", update.DateStart)
	}

	if update.DateEnd != "" && !update.DateEnd.IsValid() {
		return "", fmt.Errorf("New DateEnd Is Invalid %v", update.DateEnd)
	}

	if update.DateStart != "" && update.DateEnd != "" && update.DateStart.GetMS() > update.DateEnd.GetMS() {
		return "", fmt.Errorf("Date Start %v Is After Date End %v", update.DateStart, update.DateEnd)
	}

	return id, nil
}
