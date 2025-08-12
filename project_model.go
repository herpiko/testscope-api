package main

import (
	"log"
)

type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	InviteCode  string `json:"inviteCode"`
	Access      string `json:"access"`
	AuthorName  string `json:"authorName"`
	CreatedAt   string `json:"createdAt"`
}

type Collaborator struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	EmailAddress string `json:"emailAddress"`
	Access       string `json:"access"`
	CreatedAt    string `json:"createdAt"`
}

type Projects struct {
	Data  []Project `json:"data"`
	Page  int32     `json:"page"`
	Limit int32     `json:"limit"`
}

type Collaborators struct {
	Data []Collaborator `json:"data"`
}

func (p *Project) getProject() error {
	return app.DB.QueryRow(`
	SELECT name, description, invite_code FROM projects WHERE id=$1
	AND deleted_at IS NULL
	`,
		p.ID).Scan(&p.Name, &p.Description, &p.InviteCode)
}

func (p *Project) deleteProject() error {
	_, err := app.DB.Exec(`
	UPDATE projects SET deleted_at=NOW() WHERE id=$1;
	`, p.ID)
	if err != nil {
		return err
	}
	_, err = app.DB.Exec(`
	UPDATE scopes SET deleted_at=NOW() WHERE project_id=$1;
	`, p.ID)
	if err != nil {
		return err
	}
	_, err = app.DB.Exec(`
	UPDATE scenarios SET deleted_at=NOW() WHERE project_id=$1;
	`, p.ID)
	return err
}

func (p *Project) createProject() error {
	err := app.DB.QueryRow(`
	INSERT INTO projects(name, description) VALUES($1, $2) RETURNING id
  `,
		p.Name, p.Description).Scan(&p.ID)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func getProjects(start, count int, userId string) ([]Project, error) {
	rows, err := app.DB.Query(`
	SELECT projects.id, projects.name, projects.description, projects.created_at, users.email_address
	FROM projects, access_control_lists acl, users
	WHERE projects.deleted_at IS NULL AND
	projects.id::text=acl.object_id::text AND acl.object_type='project' AND acl.user_id=$3
	AND acl.user_id::text = users.id::text
	GROUP BY projects.id, users.email_address
	ORDER BY projects.created_at ASC
	LIMIT $1 OFFSET $2
  `,
		count, start, userId)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()

	projects := []Project{}

	for rows.Next() {
		var p Project
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.CreatedAt,
			&p.AuthorName,
		); err != nil {
			log.Println(err)
			return nil, err
		}
		projects = append(projects, p)
	}

	return projects, nil
}

func (p *Project) getInvitation() error {
	return app.DB.QueryRow(`
	SELECT id, name, invite_code FROM projects WHERE invite_code=$1
	AND deleted_at IS NULL
	`,
		p.InviteCode).Scan(&p.ID, &p.Name, &p.InviteCode)
}

func (p *Project) updateProject() error {
	_, err :=
		app.DB.Exec(`
		UPDATE projects SET name=$1, description=$2, updated_at=NOW()
		WHERE id=$3
		`,
			p.Name, p.Description, p.ID)

	return err
}

func getCollaborators(projectId string) (*Collaborators, error) {
	rows, err := app.DB.Query(`
	SELECT users.id, users.user_name, users.email_address, acl.access, max(acl.created_at)
	FROM users, access_control_lists acl WHERE
	users.id::text=acl.user_id::text AND acl.object_id=$1
	GROUP BY users.id, acl.access
	ORDER BY max(acl.created_at)
  `,
		projectId)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()

	collaborators := []Collaborator{}

	for rows.Next() {
		var p Collaborator
		if err := rows.Scan(
			&p.ID,
			&p.Username,
			&p.EmailAddress,
			&p.Access,
			&p.CreatedAt,
		); err != nil {
			log.Println(err)
			return nil, err
		}
		collaborators = append(collaborators, p)
	}

	res := Collaborators{}
	res.Data = collaborators

	return &res, nil
}
