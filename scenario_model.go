package main

import (
	"database/sql"
	"encoding/json"
	"log"
)

type Scenario struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	ScopeID   string `json:"scopeId"`
	Name      string `json:"name"`
	Steps     []Step `json:"steps"`

	// Optional
	AssigneeID   string   `json:"assigneeId"`
	AssigneeName string   `json:"assigneeName"`
	Assists      []Assist `json:"assists"`
	Status       int      `json:"status"`
	Notes        string   `json:"notes"`
}

type Step struct {
	Step        string `json:"step"`
	Expectation string `json:"expectation"`
	Passed      bool   `json:"passed"`
}

type Scenarios struct {
	Data  []Scenario `json:"data"`
	Page  int32      `json:"page"`
	Limit int32      `json:"limit"`
}

func (p *Scenario) getScenario() error {
	var steps sql.NullString
	err := app.DB.QueryRow(`
	SELECT name, scope_id, project_id, steps FROM scenarios WHERE id=$1
	AND deleted_at IS NULL
	`,
		p.ID).Scan(&p.Name, &p.ScopeID, &p.ProjectID, &steps)
	if err != nil {
		log.Println(err)
		return err
	}

	err = json.Unmarshal([]byte(steps.String), &p.Steps)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (p *Scenario) updateScenario() error {
	jsonBytes, _ := json.Marshal(p.Steps)
	_, err :=
		app.DB.Exec(`
		UPDATE scenarios SET name=$1, steps=$2, scope_id=$3, updated_at=NOW()
		WHERE id=$4
		`,
			p.Name, string(jsonBytes), p.ScopeID, p.ID)

	return err
}

func (p *Scenario) deleteScenario() error {
	_, err := app.DB.Exec(`
	UPDATE scenarios SET deleted_at=NOW() WHERE id=$1
	`, p.ID)
	return err
}

func (p *Scenario) createScenario() error {
	jsonBytes, _ := json.Marshal(p.Steps)
	err := app.DB.QueryRow(`
	INSERT INTO scenarios(name, scope_id, project_id, steps) VALUES($1, $2, $3, $4) RETURNING id
  `,
		p.Name, p.ScopeID, p.ProjectID, string(jsonBytes)).Scan(&p.ID)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func getScenarios(start, count int, projectId string) ([]Scenario, error) {
	rows, err := app.DB.Query(`
	SELECT id, name, scope_id, project_id FROM scenarios
	WHERE deleted_at IS NULL AND project_id=$3
	ORDER BY name ASC
	LIMIT $1 OFFSET $2
  `,
		count, start, projectId)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()

	scenarios := []Scenario{}

	for rows.Next() {
		var p Scenario
		if err := rows.Scan(&p.ID, &p.Name, &p.ScopeID, &p.ProjectID); err != nil {
			log.Println(err)
			return nil, err
		}
		scenarios = append(scenarios, p)
	}

	return scenarios, nil
}

func getScenariosBySession(start, count int, sessionID string) ([]Scenario, error) {
	rows, err := app.DB.Query(`
	SELECT scen.id, scen.name, scen.scope_id, scen.project_id
	FROM scenarios scen, sessions s WHERE scen.id::text=ANY( s.scenarios) AND s.id::text=$3
	AND scen.deleted_at IS NULL
	LIMIT $1 OFFSET $2
  `,
		count, start, sessionID)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()

	scenarios := []Scenario{}

	for rows.Next() {
		var p Scenario
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.ScopeID,
			&p.ProjectID,
		); err != nil {
			log.Println(err)
			return nil, err
		}
		scenarios = append(scenarios, p)
	}

	return scenarios, nil
}
