package main

import (
	"log"
	"strings"
)

type User struct {
	ID               string `json:"id"`
	FullName         string `json:"full_name"`
	UserName         string `json:"user_name"`
	EmailAddress     string `json:"email_address"`
	Role             string `json:"role"`
	SubscriptionType string `json:"subscription_type"`
	Quotas           Quotas `json:"quotas"`
}

type Users struct {
	Data  []User `json:"data"`
	Page  int32  `json:"page"`
	Limit int32  `json:"limit"`
}

func (p *User) getUser() error {
	return app.DB.QueryRow(`
	SELECT id, full_name, user_name, email_address, role, subscription_type
	FROM users WHERE (id::text=$1 OR email_address=$2)
	AND deleted_at IS NULL
	`,
		p.ID, p.EmailAddress).Scan(
		&p.ID,
		&p.FullName,
		&p.UserName,
		&p.EmailAddress,
		&p.Role,
		&p.SubscriptionType,
	)
}

func (p *User) updateUser() error {
	_, err :=
		app.DB.Exec(`
		UPDATE users SET
		full_name=$1,
		user_name=$2,
		email_address=$3,
		updated_at=NOW()
		WHERE id=$4`,
			p.FullName,
			p.UserName,
			p.EmailAddress,
			p.ID,
		)

	return err
}

func (p *User) deleteUser() error {
	_, err := app.DB.Exec(`
	UPDATE users SET deleted_at=NOW() WHERE id=$1
	`, p.ID)
	return err
}

func (p *User) createUser() error {
	err := app.DB.QueryRow(`
	INSERT INTO users(full_name, user_name, email_address)
	VALUES($1, $2, $3) RETURNING id, role
	`,
		p.FullName,
		p.UserName,
		p.EmailAddress,
	).Scan(&p.ID, &p.Role)

	if err != nil {
		log.Println(err)
		if strings.Contains(err.Error(), "duplicate") {
			err = nil
			u := User{}
			u.EmailAddress = p.EmailAddress
			err = u.getUser()
			if err != nil {
				log.Println(err)
				return err
			}
			p.ID = u.ID
			p.Role = u.Role
			log.Println(p)
		} else {
			return err
		}
	}

	return nil
}

func getUsers(start, count int) ([]User, error) {
	rows, err := app.DB.Query(`
	SELECT id, full_name, user_name, email_address
	FROM users
	WHERE deleted_at IS NULL
	LIMIT $1 OFFSET $2
	`,
		count, start)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()

	users := []User{}

	for rows.Next() {
		var p User
		if err := rows.Scan(
			&p.ID,
			&p.FullName,
			&p.UserName,
			&p.EmailAddress,
		); err != nil {
			log.Println(err)
			return nil, err
		}
		users = append(users, p)
	}

	return users, nil
}
