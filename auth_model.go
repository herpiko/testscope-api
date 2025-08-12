package main

import (
	"database/sql"
	"errors"
	"log"
)

var ACL_LEVELS = [...]string{
	"OWNER",
	"MODIFY",
	"READ",
}

type Token struct {
	Key          string
	EmailAddress string
	UserID       string
	AuthProvider string
}

type Quotas struct {
	SubscriptionType string `json:"subscription_type"`
	Project          int    `json:"project"`
	Scope            int    `json:"scope"`
	Scenario         int    `json:"scenario"`
	Session          int    `json:"session"`
	Test             int    `json:"test"`
}

type Acl struct {
	ObjectID   string
	ObjectType string
	UserID     string
	Access     string
}

type ParentChilds struct {
	Parent string
	Childs []string
}

func (t *Token) verifyCachedToken() error {
	return app.DB.QueryRow(`
	SELECT user_id, email_address, auth_provider FROM tokens WHERE key=$1
	AND deleted_at IS NULL
	`,
		t.Key).Scan(
		&t.UserID,
		&t.EmailAddress,
		&t.AuthProvider,
	)
}

func (t *Token) cacheToken() error {
	_, err := app.DB.Exec(`
	INSERT INTO tokens (key, user_id, email_address, auth_provider) VALUES ($1, $2, $3, $4)
	`,
		t.Key, t.UserID, t.EmailAddress, t.AuthProvider)
	return err
}

func (ac *Acl) createAccess() error {
	isValid := false
	for _, level := range ACL_LEVELS {
		if level == ac.Access {
			isValid = true
			break
		}
	}
	if !isValid {
		err := errors.New("invalid-access-level")
		log.Println(err)
		return err
	}
	_, err := app.DB.Exec(`
	INSERT INTO access_control_lists
	(object_id, object_type, user_id, access)
	VALUES ($1, $2, $3, $4)
	`,
		ac.ObjectID, ac.ObjectType, ac.UserID, ac.Access)
	return err
}

func (ac *Acl) dropAccess() error {
	_, err := app.DB.Exec(`
	DELETE FROM access_control_lists
	WHERE object_id=$1 AND object_type=$2 AND user_id=$3
	`,
		ac.ObjectID, ac.ObjectType, ac.UserID)
	return err
}

func (a *Acl) getAccess() error {
	return app.DB.QueryRow(`
    SELECT object_type, access
    FROM access_control_lists
    WHERE object_id=$1 AND user_id=$2
	`,
		a.ObjectID,
		a.UserID,
	).Scan(
		&a.ObjectType,
		&a.Access,
	)
}

func (p *ParentChilds) createParentChilds() error {
	var err error
	tx, err := app.DB.Begin()
	defer tx.Rollback()
	for _, child := range p.Childs {
		_, err := tx.Exec(`
			INSERT INTO parent_childs (parent, child) VALUES ($1, $2)
			`,
			p.Parent, child)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return err
	}
	return err
}

func getUserQuotas(userID string) (Quotas, error) {
	var quotas Quotas

	// Project
	err := app.DB.QueryRow(`
	SELECT users.subscription_type, COUNT(*) FROM projects p, access_control_lists acl, users WHERE p.id::text=acl.object_id AND users.id::text=acl.user_id AND object_type='project' AND p.deleted_at IS NULL AND users.id=$1 GROUP BY users.id;
	`,
		userID).Scan(
		&quotas.SubscriptionType,
		&quotas.Project,
	)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return quotas, err
	}

	// Scope
	err = app.DB.QueryRow(`
	SELECT COUNT(scopes.id) FROM scopes, projects, access_control_lists acl, users WHERE scopes.project_id::text=projects.id::text AND acl.object_id=projects.id::text AND acl.access='OWNER' AND acl.user_id::text=users.id::text AND users.id=$1 AND scopes.deleted_at IS NULL GROUP BY users.id;
	`, userID).Scan(
		&quotas.Scope,
	)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return quotas, err
	}

	// Scenario
	err = app.DB.QueryRow(`
	SELECT COUNT(scenarios.id) FROM scenarios, projects, access_control_lists acl, users WHERE scenarios.project_id::text=projects.id::text AND acl.object_id=projects.id::text AND acl.access='OWNER' AND acl.user_id::text=users.id::text AND users.id=$1 AND scenarios.deleted_at IS NULL GROUP BY users.id;
	`,
		userID).Scan(
		&quotas.Scenario,
	)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return quotas, err
	}

	// Session
	err = app.DB.QueryRow(`
	SELECT COUNT(sessions.id) FROM sessions, projects, access_control_lists acl, users WHERE sessions.project_id::text=projects.id::text AND acl.object_id=projects.id::text AND acl.access='OWNER' AND acl.user_id::text=users.id::text AND users.id=$1 AND sessions.deleted_at IS NULL GROUP BY users.id;
	`,
		userID).Scan(
		&quotas.Session,
	)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return quotas, err
	}

	// Test
	err = app.DB.QueryRow(`
SELECT COUNT(tests.id) FROM tests, projects, sessions, access_control_lists acl, users WHERE tests.session_id::text=sessions.id::text AND sessions.project_id::text=projects.id::text AND acl.object_id=projects.id::text AND acl.access='OWNER' AND acl.user_id::text=users.id::text AND users.id=$1 AND sessions.deleted_at IS NULL GROUP BY users.id;
	`,
		userID).Scan(
		&quotas.Test,
	)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return quotas, err
	}

	return quotas, nil

}

func isEligibleToCreateProject(userID string) (bool, error) {
	var subscriptionType string
	var count int
	err := app.DB.QueryRow(`
	SELECT users.subscription_type, COUNT(*) FROM projects p, access_control_lists acl, users WHERE p.id::text=acl.object_id AND users.id::text=acl.user_id AND object_type='project' AND p.deleted_at IS NULL AND users.id=$1 GROUP BY users.id;
	`,
		userID).Scan(
		&subscriptionType,
		&count,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return true, nil
		}
		log.Println(err)
		return false, err
	}
	if subscriptionType == "free" && count >= 3 {
		return false, nil
	}
	if subscriptionType == "standard" && count >= 10 {
		return false, nil
	}
	return true, nil

}

func isEligibleToCreateScope(projectID string) (bool, error) {
	var userID string
	err := app.DB.QueryRow(`
    SELECT users.id FROM projects p, access_control_lists acl, users WHERE p.id::text=acl.object_id AND users.id::text=acl.user_id AND object_type='project' AND acl.access='OWNER' AND p.deleted_at IS NULL
    AND p.id=$1
	`,
		projectID).Scan(
		&userID,
	)
	if err != nil {
		log.Println(err)
		return false, err
	}

	var count int
	var subscriptionType string
	err = app.DB.QueryRow(`
	SELECT users.subscription_type, COUNT(scopes.id) FROM scopes, projects, access_control_lists acl, users WHERE scopes.project_id::text=projects.id::text AND acl.object_id=projects.id::text AND acl.access='OWNER' AND acl.user_id::text=users.id::text AND users.id=$1 AND scopes.deleted_at IS NULL GROUP BY users.id;
	`,
		userID).Scan(
		&subscriptionType,
		&count,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return true, nil
		}
		log.Println(err)
		return false, err
	}
	if subscriptionType == "free" && count >= 10 {
		return false, nil
	}
	if subscriptionType == "standard" && count >= 100 {
		return false, nil
	}
	return true, nil
}

func isEligibleToCreateScenario(projectID string) (bool, error) {
	var userID string
	err := app.DB.QueryRow(`
    SELECT users.id FROM projects p, access_control_lists acl, users WHERE p.id::text=acl.object_id AND users.id::text=acl.user_id AND object_type='project' AND p.deleted_at IS NULL
    AND p.id=$1
	`,
		projectID).Scan(
		&userID,
	)
	if err != nil {
		log.Println(err)
		return false, err
	}

	var subscriptionType string
	var count int
	err = app.DB.QueryRow(`
	SELECT users.subscription_type, COUNT(scenarios.id) FROM scenarios, projects, access_control_lists acl, users WHERE scenarios.project_id::text=projects.id::text AND acl.object_id=projects.id::text AND acl.access='OWNER' AND acl.user_id::text=users.id::text AND users.id=$1 AND scenarios.deleted_at IS NULL GROUP BY users.id;
	`,
		userID).Scan(
		&subscriptionType,
		&count,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return true, nil
		}
		log.Println(err)
		return false, err
	}
	if subscriptionType == "free" && count >= 50 {
		return false, nil
	}
	if subscriptionType == "standard" && count >= 1000 {
		return false, nil
	}
	return true, nil
}

func isEligibleToCreateSession(projectID string) (bool, error) {
	var userID string
	err := app.DB.QueryRow(`
    SELECT users.id FROM projects p, access_control_lists acl, users WHERE p.id::text=acl.object_id AND users.id::text=acl.user_id AND object_type='project' AND p.deleted_at IS NULL
    AND p.id=$1
	`,
		projectID).Scan(
		&userID,
	)
	if err != nil {
		log.Println(err)
		return false, err
	}

	var subscriptionType string
	var count int
	err = app.DB.QueryRow(`
	SELECT users.subscription_type, COUNT(sessions.id) FROM sessions, projects, access_control_lists acl, users WHERE sessions.project_id::text=projects.id::text AND acl.object_id=projects.id::text AND acl.access='OWNER' AND acl.user_id::text=users.id::text AND users.id=$1 AND sessions.deleted_at IS NULL GROUP BY users.id;
	`,
		userID).Scan(
		&subscriptionType,
		&count,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return true, nil
		}
		log.Println(err)
		return false, err
	}
	if subscriptionType == "free" && count >= 50 {
		return false, nil
	}
	if subscriptionType == "standard" && count >= 1000 {
		return false, nil
	}
	return true, nil
}
