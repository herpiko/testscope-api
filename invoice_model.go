package main

import (
	"encoding/json"
	"log"
	"time"
)

type Invoice struct {
	ID           string    `json:"id"`
	ExternalID   string    `json:"external_id"`
	UserID       string    `json:"user_id"`
	EmailAddress string    `json:"email_address"`
	Description  string    `json:"description"`
	URL          string    `json:"url"`
	Amount       float64   `json:"amount"`
	Status       string    `json:"status"`
	Items        []Product `json:"items"`
	// Populated on xendit callback event
	PaidAmount         float64   `json:"paid_amount"`
	PaymentMethod      string    `json:"payment_method"`
	PaymentChannel     string    `json:"Payment_channel"`
	PaymentDestination string    `json:"Payment_destination"`
	PaidAt             time.Time `json:"paid_at"`
	SubscriptionType   string    `json:"subscription_type"`
}

func (i *Invoice) getInvoice() error {
	return app.DB.QueryRow(`
	SELECT 
	id, external_id, user_id, email_address, description, url, amount, status
	FROM invoices
	WHERE (id::text=$1 OR external_id=$2 OR user_id=$3)
	AND deleted_at IS NULL AND created_at >= NOW() - INTERVAL '24 HOURS'
	ORDER BY created_at DESC
	`,
		i.ID, i.ExternalID, i.UserID).Scan(
		&i.ID,
		&i.ExternalID,
		&i.UserID,
		&i.EmailAddress,
		&i.Description,
		&i.URL,
		&i.Amount,
		&i.Status,
	)
}

func (i *Invoice) updateInvoice() error {
	tx, err := app.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err =
		tx.QueryRow(`
		UPDATE invoices SET
		status=$3,
		paid_amount=$4,
		payment_method=$5,
		payment_channel=$6,
		payment_destination=$7,
		paid_at=NOW(),
		updated_at=NOW()
	  WHERE (id::text=$1 OR external_id=$2)
	  RETURNING user_id, description
		`,
			i.ID,
			i.ExternalID,
			i.Status,
			i.PaidAmount,
			i.PaymentMethod,
			i.PaymentChannel,
			i.PaymentDestination,
		).Scan(&i.UserID, &i.SubscriptionType)
	if err != nil {
		return err
	}
	_, err =
		tx.Exec(`
		UPDATE users SET
		subscription_type=$2
	  WHERE id::text=$1
		`,
			i.UserID,
			i.SubscriptionType,
		)
	if err != nil {
		return err
	}

	tx.Commit()

	return err
}

func (i *Invoice) deleteInvoice() error {
	_, err := app.DB.Exec(`
	UPDATE invoices SET deleted_at=NOW() 
	WHERE (id::text=$1 OR external_id=$2)
	`, i.ID)
	return err
}

func (i *Invoice) createInvoice() error {
	jsonByte, _ := json.Marshal(i.Items)
	items := string(jsonByte)
	err := app.DB.QueryRow(`
	INSERT INTO 
	invoices(id, external_id, user_id, email_address, description, url, amount, status, items)
	VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id
	`,
		i.ID,
		i.ExternalID,
		i.UserID,
		i.EmailAddress,
		i.Description,
		i.URL,
		i.Amount,
		i.Status,
		items,
	).Scan(&i.ID)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func getInvoices(start, count int) ([]Invoice, error) {
	rows, err := app.DB.Query(`
	SELECT
	id, external_id, user_id, email_address, description, url, amount, status
	FROM invoices
	WHERE deleted_at IS NULL
	LIMIT $1 OFFSET $2
	`,
		count, start)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()

	invoices := []Invoice{}

	for rows.Next() {
		var i Invoice
		if err := rows.Scan(
			&i.ID,
			&i.ExternalID,
			&i.UserID,
			&i.EmailAddress,
			&i.Description,
			&i.URL,
			&i.Amount,
			&i.Status,
		); err != nil {
			log.Println(err)
			return nil, err
		}
		invoices = append(invoices, i)
	}

	return invoices, nil
}
