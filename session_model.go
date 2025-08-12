package main

import (
	"database/sql"
	"encoding/json"
	"log"

	"github.com/lib/pq"
)

type Session struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"projectId"`
	AuthorID    string     `json:"authorId"`
	AuthorName  string     `json:"authorName"`
	Version     string     `json:"version"`
	Description string     `json:"description"`
	Status      int        `json:"status"`
	Scenarios   []Scenario `json:"scenarios"`
	CreatedAt   string     `json:"createdAt"`
}

type Assist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Test struct {
	ID           string   `json:"id"`
	SessionID    string   `json:"sessionId"`
	AssigneeID   string   `json:"assigneeId"`
	AssigneeName string   `json:"assigneeName"`
	ScenarioID   string   `json:"scenarioId"`
	Steps        []Step   `json:"steps"`
	Status       int      `json:"status"`
	Notes        string   `json:"notes"`
	CreatedAt    string   `json:"createdAt"`
	Assists      []Assist `json:"assists"`
}

type Sessions struct {
	Data  []Session `json:"data"`
	Page  int32     `json:"page"`
	Limit int32     `json:"limit"`
}

func (p *Session) getSession() error {
	err := app.DB.QueryRow(`
	SELECT id, project_id, author_id, version, description, status
	FROM sessions WHERE id=$1
	AND deleted_at IS NULL
	`,
		p.ID).Scan(
		&p.ID,
		&p.ProjectID,
		&p.AuthorID,
		&p.Version,
		&p.Description,
		&p.Status,
	)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil

}

func (p *Session) deleteSession() error {
	_, err := app.DB.Exec(`
	UPDATE sessions SET deleted_at=NOW() WHERE id=$1
	`, p.ID)
	return err
}

func (p *Session) resetSession() error {
	_, err := app.DB.Exec(`
	UPDATE tests SET status=3, deleted_at=NOW() WHERE session_id=$1
	`, p.ID)
	return err
}

func (p *Session) createSession() error {
	arr := []string{}
	for _, scen := range p.Scenarios {
		arr = append(arr, scen.ID)
	}

	err := app.DB.QueryRow(`
	INSERT INTO sessions(
	  project_id,
	  author_id,
	  version,
	  description,
	  scenarios
	) VALUES($1, $2, $3, $4, $5) RETURNING id
  `,
		p.ProjectID,
		p.AuthorID,
		p.Version,
		p.Description,
		pq.Array(arr),
	).Scan(&p.ID)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func getSessions(start, count int, projectId string) ([]Session, error) {
	rows, err := app.DB.Query(`
	SELECT 
	sessions.id,
	sessions.project_id,
	sessions.author_id,
	sessions.version,
	sessions.description,
	sessions.status,
	sessions.scenarios,
	sessions.created_at
	FROM sessions
	WHERE sessions.deleted_at IS NULL
	AND sessions.project_id=$3
	AND sessions.status != 3
	ORDER BY sessions.created_at ASC
	LIMIT $1 OFFSET $2
  `,
		count, start, projectId)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()

	sessions := []Session{}
	arr := []string{}

	for rows.Next() {
		var p Session
		if err := rows.Scan(
			&p.ID,
			&p.ProjectID,
			&p.AuthorID,
			&p.Version,
			&p.Description,
			&p.Status,
			pq.Array(&arr),
			&p.CreatedAt,
		); err != nil {
			log.Println(err)
			return nil, err
		}
		p.Scenarios = []Scenario{}
		for _, scenID := range arr {
			p.Scenarios = append(p.Scenarios, Scenario{ID: scenID})
		}
		sessions = append(sessions, p)
	}

	return sessions, nil
}

func getTests(start, count int, sessionId string) ([]Test, error) {
	rows, err := app.DB.Query(`
  SELECT t.id, u.id, u.email_address, t.scenario_id, t.steps, t.status, t.notes, t.created_at, t.assists
	FROM users u, tests t WHERE u.id = t.assignee_id AND t.session_id::text=$3 AND t.deleted_at IS NULL
	LIMIT $1 OFFSET $2
  `,
		count, start, sessionId)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()

	tests := []Test{}

	for rows.Next() {
		var p Test
		var steps sql.NullString
		assists := []string{}
		if err := rows.Scan(
			&p.ID,
			&p.AssigneeID,
			&p.AssigneeName,
			&p.ScenarioID,
			&steps,
			&p.Status,
			&p.Notes,
			&p.CreatedAt,
			pq.Array(&assists),
		); err != nil {
			log.Println(err)
			return nil, err
		}

		for _, assist := range assists {
			p.Assists = append(p.Assists, Assist{ID: assist})
		}

		err = json.Unmarshal([]byte(steps.String), &p.Steps)
		if err != nil {
			log.Println(err)
		}
		tests = append(tests, p)
	}

	return tests, nil
}

func (p *Session) updateSession() error {
	arr := []string{}
	for _, scen := range p.Scenarios {
		arr = append(arr, scen.ID)
	}
	_, err :=
		app.DB.Exec(`
		UPDATE sessions SET
		version=$1,
		description=$2,
		status=$3,
		scenarios=$4,
		updated_at=NOW()
		WHERE id=$5
		`,
			p.Version,
			p.Description,
			p.Status,
			pq.Array(arr),
			p.ID,
		)

	return err
}

func (p *Test) createTest() error {
	_, err := app.DB.Exec(`
	UPDATE tests SET status=3, deleted_at=NOW()
	WHERE assignee_id=$1 AND scenario_id=$2 AND session_id=$3
	`,
		p.AssigneeID,
		p.ScenarioID,
		p.SessionID,
	)
	if err != nil {
		log.Println(err)
		return err
	}

	jsonBytes, _ := json.Marshal(p.Steps)
	log.Println(string(jsonBytes))
	err = app.DB.QueryRow(`
	INSERT INTO tests(
	  session_id,
	  assignee_id,
	  scenario_id,
	  steps,
	  status
	) VALUES($1, $2, $3, $4, $5) RETURNING id, created_at
  `,
		p.SessionID,
		p.AssigneeID,
		p.ScenarioID,
		string(jsonBytes),
		p.Status, // on going
	).Scan(&p.ID, &p.CreatedAt)

	if err != nil {
		log.Println(err)
		return err
	}

	scen := Scenario{ID: p.ScenarioID}
	err = scen.getScenario()
	if err != nil {
		log.Println(err)
		return err
	}

	p.Steps = scen.Steps

	return nil
}

func (p *Test) getTestByOther() error {
	var steps sql.NullString
	assists := []string{}
	err := app.DB.QueryRow(`
  SELECT t.id, u.id, u.email_address, t.scenario_id, t.steps, t.status, t.notes, t.created_at, t.assists 
	FROM users u, tests t WHERE u.id = t.assignee_id AND t.scenario_id::text=$1
	AND u.id!=$2 AND t.status=$3 AND t.session_id=$4 AND t.deleted_at IS NULL
	`,
		p.ScenarioID,
		p.AssigneeID,
		p.Status,
		p.SessionID,
	).Scan(
		&p.ID,
		&p.AssigneeID,
		&p.AssigneeName,
		&p.ScenarioID,
		&steps,
		&p.Status,
		&p.CreatedAt,
		&p.CreatedAt,
		pq.Array(&assists),
	)
	for _, assist := range assists {
		p.Assists = append(p.Assists, Assist{ID: assist})
	}

	err = json.Unmarshal([]byte(steps.String), &p.Steps)
	if err != nil {
		log.Println(err)
	}
	return nil
}

func (p *Test) getTestByAssignee() error {
	var steps sql.NullString
	assists := []string{}
	err := app.DB.QueryRow(`
  SELECT t.id, u.id, u.email_address, t.scenario_id, t.steps, t.status, t.notes, t.created_at, t.assists 
	FROM users u, tests t WHERE u.id = t.assignee_id AND t.scenario_id::text=$1
	AND u.id=$2 AND t.status=$3 AND t.session_id=$4 AND t.status != 3
	`,
		p.ScenarioID,
		p.AssigneeID,
		p.Status,
		p.SessionID,
	).Scan(
		&p.ID,
		&p.AssigneeID,
		&p.AssigneeName,
		&p.ScenarioID,
		&steps,
		&p.Status,
		&p.CreatedAt,
		&p.CreatedAt,
		pq.Array(&assists),
	)

	for _, assist := range assists {
		p.Assists = append(p.Assists, Assist{ID: assist})
	}

	err = json.Unmarshal([]byte(steps.String), &p.Steps)
	if err != nil {
		log.Println(err)
	}
	return nil
}

func (p *Test) deleteTest() error {
	_, err := app.DB.Exec(`
	UPDATE tests SET status=3, deleted_at=NOW() WHERE id=$1
	`, p.ID)
	return err
}

func (p *Test) updateTest() error {
	assists := []string{}
	for _, assist := range p.Assists {
		assists = append(assists, assist.ID)
	}
	jsonBytes, _ := json.Marshal(p.Steps)
	log.Println(string(jsonBytes))
	_, err :=
		app.DB.Exec(`
		UPDATE tests SET
		steps=$1,
		status=$2,
		notes=$3,
		assists=$4,
		updated_at=NOW()
		WHERE id=$5
		`,
			string(jsonBytes),
			p.Status,
			p.Notes,
			pq.Array(assists),
			p.ID,
		)

	return err
}
