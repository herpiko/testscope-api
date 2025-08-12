package main

type Product struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

func (p *Product) getProduct() error {
	return app.DB.QueryRow(`
	SELECT 
	id, name, amount
	FROM products
	WHERE id::text=$1
	AND deleted_at IS NULL
	`,
		p.ID).Scan(
		&p.ID,
		&p.Name,
		&p.Amount,
	)
}
