package main

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlobUploadFile(t *testing.T) {

	app.MigrateClean()
	defer app.DB.Close()

	req, _ := http.NewRequest("GET", "/", nil)
	response := executeRequest(req)
	assert.Equal(t, http.StatusOK, response.Code)

	// Using token is also ok
	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", testUserToken1)
	response = executeRequest(req)
	assert.Equal(t, http.StatusOK, response.Code)
}
