package main

import (
	"fmt"
	"testing"

	"bytes"
	"encoding/json"
	"net/http"

	uuidParser "github.com/docker/distribution/uuid"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestUserEmptyTable(t *testing.T) {
	app.MigrateClean()
	defer app.DB.Close()

	req, _ := http.NewRequest("GET", "/api/users", nil)
	req.Header.Set("Authorization", testUserToken1)
	response := executeRequest(req)

	assert.Equal(t, http.StatusOK, response.Code)

	body := response.Body.String()
	assert.Equal(t, true, len(body) > 0) // Already populated by auth
}

func TestUserGetNonExistent(t *testing.T) {
	app.MigrateClean()
	defer app.DB.Close()

	req, _ := http.NewRequest("GET", "/api/user/"+uuid.NewV4().String(), nil)
	req.Header.Set("Authorization", testUserToken1)
	response := executeRequest(req)

	// TODO should be 404, need to fix it in middleware
	assert.Equal(t, http.StatusForbidden, response.Code)
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	assert.Equal(t, "forbidden", fmt.Sprintf("%s", m["error"]))
}

func TestUserCreate(t *testing.T) {
	app.MigrateClean()
	defer app.DB.Close()

	jsonStr := []byte(`{"full_name":"Asep Indrayana", "user_name": "asep", "email_address": "asep@testscope.io"}`)
	req, _ := http.NewRequest("POST", "/api/user", bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", testUserToken1)
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)
	assert.Equal(t, http.StatusCreated, response.Code)
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	fullName := fmt.Sprintf("%s", m["full_name"])
	assert.Equal(t, "Asep Indrayana", fullName)
	userName := fmt.Sprintf("%s", m["user_name"])
	assert.Equal(t, "asep", userName)
	emailAddress := fmt.Sprintf("%s", m["email_address"])
	assert.Equal(t, "asep@testscope.io", emailAddress)
	id := fmt.Sprintf("%s", m["id"])
	_, err := uuidParser.Parse(id)
	assert.Equal(t, nil, err)

	// Fail to create with the same email address
	// because of unique constraint
	jsonStr = []byte(`{"full_name":"Asep Indrayana", "user_name": "asep", "email_address": "asep@testscope.io"}`)
	req, _ = http.NewRequest("POST", "/api/user", bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", testUserToken1)
	req.Header.Set("Content-Type", "application/json")
	response = executeRequest(req)
	assert.Equal(t, http.StatusInternalServerError, response.Code)

}

func TestUserGet(t *testing.T) {
	app.MigrateClean()
	defer app.DB.Close()

	jsonStr := []byte(`{"full_name":"Asep Indrayana", "user_name": "asep", "email_address": "asep@testscope.io"}`)
	req, _ := http.NewRequest("POST", "/api/user", bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", testUserToken1)
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)
	assert.Equal(t, http.StatusCreated, response.Code)
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/user/%s", m["id"]), nil)
	req.Header.Set("Authorization", testUserToken1)
	response = executeRequest(req)
	assert.Equal(t, http.StatusOK, response.Code)
	json.Unmarshal(response.Body.Bytes(), &m)
	assert.Equal(t, "asep@testscope.io", fmt.Sprintf("%s", m["email_address"]))

	req, _ = http.NewRequest("GET", "/api/user", nil)
	req.Header.Set("Authorization", testUserToken1)
	response = executeRequest(req)
	assert.Equal(t, http.StatusOK, response.Code)
	json.Unmarshal(response.Body.Bytes(), &m)
	assert.Equal(t, "padfoot.tgz@gmail.com", fmt.Sprintf("%s", m["email_address"]))
}

func TestUserUpdate(t *testing.T) {
	app.MigrateClean()
	defer app.DB.Close()

	// Add item
	jsonStr := []byte(`{"full_name":"Asep Indrayana", "user_name": "asep", "email_address": "asep@testscope.io"}`)
	req, _ := http.NewRequest("POST", "/api/user", bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", testUserToken1)
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)
	assert.Equal(t, http.StatusCreated, response.Code)
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	fullName := fmt.Sprintf("%s", m["full_name"])
	assert.Equal(t, "Asep Indrayana", fullName)
	userName := fmt.Sprintf("%s", m["user_name"])
	assert.Equal(t, "asep", userName)
	emailAddress := fmt.Sprintf("%s", m["email_address"])
	assert.Equal(t, "asep@testscope.io", emailAddress)
	id := fmt.Sprintf("%s", m["id"])
	_, err := uuidParser.Parse(id)
	assert.Equal(t, nil, err)

	// Get the original
	req, _ = http.NewRequest("GET", "/api/user/"+id, nil)
	req.Header.Set("Authorization", testUserToken1)
	response = executeRequest(req)
	var originalUser map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalUser)

	// Update
	jsonStr = []byte(`{"full_name":"M Asep Indrayana", "user_name": "asep", "email_address": "asep@testscope.io"}`)
	req, _ = http.NewRequest("PUT", "/api/user/"+id, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", testUserToken1)
	req.Header.Set("Content-Type", "application/json")
	response = executeRequest(req)
	assert.Equal(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/api/user/"+id, nil)
	req.Header.Set("Authorization", testUserToken1)
	response = executeRequest(req)
	var updated map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &updated)
	assert.Equal(t, originalUser["id"], updated["id"])
	assert.NotEqual(t, originalUser["full_name"], updated["full_name"])
	assert.Equal(t, "M Asep Indrayana", fmt.Sprintf("%s", updated["full_name"]))
}

func TestUserDelete(t *testing.T) {
	app.MigrateClean()
	defer app.DB.Close()

	// Add item
	jsonStr := []byte(`{"full_name":"Asep Indrayana", "user_name": "asep", "email_address": "asep@testscope.io"}`)
	req, _ := http.NewRequest("POST", "/api/user", bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", testUserToken1)
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)
	assert.Equal(t, http.StatusCreated, response.Code)
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	fullName := fmt.Sprintf("%s", m["full_name"])
	assert.Equal(t, "Asep Indrayana", fullName)
	userName := fmt.Sprintf("%s", m["user_name"])
	assert.Equal(t, "asep", userName)
	emailAddress := fmt.Sprintf("%s", m["email_address"])
	assert.Equal(t, "asep@testscope.io", emailAddress)
	id := fmt.Sprintf("%s", m["id"])
	_, err := uuidParser.Parse(id)
	assert.Equal(t, nil, err)

	req, _ = http.NewRequest("GET", "/api/user/"+id, nil)
	req.Header.Set("Authorization", testUserToken1)
	response = executeRequest(req)
	assert.Equal(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/api/user/"+id, nil)
	req.Header.Set("Authorization", testUserToken1)
	response = executeRequest(req)
	assert.Equal(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/api/user/"+id, nil)
	req.Header.Set("Authorization", testUserToken1)
	response = executeRequest(req)
	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestUserAdminDelete(t *testing.T) {
	app.MigrateClean()
	defer app.DB.Close()

	// Add item
	jsonStr := []byte(`{"full_name":"Asep Indrayana", "user_name": "asep", "email_address": "asep@testscope.io"}`)
	req, _ := http.NewRequest("POST", "/api/user", bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", testUserToken1)
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)
	assert.Equal(t, http.StatusCreated, response.Code)
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	fullName := fmt.Sprintf("%s", m["full_name"])
	assert.Equal(t, "Asep Indrayana", fullName)
	userName := fmt.Sprintf("%s", m["user_name"])
	assert.Equal(t, "asep", userName)
	emailAddress := fmt.Sprintf("%s", m["email_address"])
	assert.Equal(t, "asep@testscope.io", emailAddress)
	id := fmt.Sprintf("%s", m["id"])
	_, err := uuidParser.Parse(id)
	assert.Equal(t, nil, err)

	req, _ = http.NewRequest("GET", "/api/user/"+id, nil)
	req.Header.Set("Authorization", testUserToken1)
	response = executeRequest(req)
	assert.Equal(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/api/user/"+id, nil)
	req.Header.Set("Authorization", testUserToken2)
	response = executeRequest(req)
	assert.Equal(t, http.StatusForbidden, response.Code)

	req, _ = http.NewRequest("DELETE", "/api/user/"+id, nil)
	req.Header.Set("Authorization", testAdminToken1)
	response = executeRequest(req)
	assert.Equal(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/api/user/"+id, nil)
	req.Header.Set("Authorization", testUserToken1)
	response = executeRequest(req)
	assert.Equal(t, http.StatusNotFound, response.Code)
}
