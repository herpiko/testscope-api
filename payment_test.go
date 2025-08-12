package main

import (
	"testing"

	"bytes"
	"net/http"

	"github.com/stretchr/testify/assert"
)

func TestPaymentCreateInvoice(t *testing.T) {
	app.MigrateClean()
	defer app.DB.Close()

	jsonStr := []byte(`{"ok": "ok"}`)
	req, _ := http.NewRequest("POST", "/api/user", bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", testUserToken1)
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)
	assert.Equal(t, http.StatusCreated, response.Code)

}
