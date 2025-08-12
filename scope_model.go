package main

import (
	"log"
)

type Scope struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"projectId"`
	ProjectName string     `json:"projectName"`
	Name        string     `json:"name"`
	Scenarios   []Scenario `json:"scenarios"`
}

type Scopes struct {
	Data  []Scope `json:"data"`
	Page  int32   `json:"page"`
	Limit int32   `json:"limit"`
}

func (p *Scope) getScope() error {
	return app.DB.QueryRow(`
	SELECT scopes.name, scopes.project_id, projects.name FROM scopes, projects WHERE scopes.id=$1
	AND scopes.project_id=projects.id
	AND scopes.deleted_at IS NULL
	`,
		p.ID).Scan(&p.Name, &p.ProjectID, &p.ProjectName)
}

func (p *Scope) updateScope() error {
	_, err :=
		app.DB.Exec(`
		UPDATE scopes SET name=$1, updated_at=NOW()
		WHERE id=$2
		`,
			p.Name, p.ID)

	return err
}

func (p *Scope) deleteScope() error {
	_, err := app.DB.Exec(`
	UPDATE scopes SET deleted_at=NOW() WHERE id=$1;
	`, p.ID)
	if err != nil {
		return err
	}
	_, err = app.DB.Exec(`
	UPDATE scenarios SET deleted_at=NOW() WHERE scope_id=$1;
	`, p.ID)
	return err
}

func (p *Scope) createScope() error {
	log.Println(p)
	err := app.DB.QueryRow(`
	INSERT INTO scopes(name, project_id) VALUES($1, $2) RETURNING id
  `,
		p.Name, p.ProjectID).Scan(&p.ID)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func getScopes(start, count int, projectId string) ([]Scope, error) {
	rows, err := app.DB.Query(`
	SELECT id, name, project_id FROM scopes
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

	scopes := []Scope{}

	for rows.Next() {
		var p Scope
		if err := rows.Scan(&p.ID, &p.Name, &p.ProjectID); err != nil {
			log.Println(err)
			return nil, err
		}
		scopes = append(scopes, p)
	}

	return scopes, nil
}
