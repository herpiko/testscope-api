package main

import (
	"net/http"
	"testing"

	uuidParser "github.com/docker/distribution/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticateIDToken(t *testing.T) {

	app.MigrateClean()
	defer app.DB.Close()

	// Verify against firebase
	currentUser, err := app.authenticateIDToken(testUserToken1)
	assert.Equal(t, nil, err)
	assert.Equal(t, "padfoot.tgz@gmail.com", currentUser.EmailAddress)
	_, err = uuidParser.Parse(currentUser.ID)
	assert.Equal(t, nil, err)

	// Using cached token
	currentUserFromCached, err := app.authenticateIDToken(testUserToken1)
	assert.Equal(t, nil, err)
	assert.Equal(t, "padfoot.tgz@gmail.com", currentUser.EmailAddress)
	_, err = uuidParser.Parse(currentUser.ID)
	assert.Equal(t, nil, err)
	assert.Equal(t, currentUser.ID, currentUserFromCached.ID)

	// Invalid jwt
	currentUser, err = app.authenticateIDToken("x")
	assert.NotEqual(t, nil, err)
}

func TestAuthPublicEndpoint(t *testing.T) {
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
