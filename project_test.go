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

func TestProjectEmptyTable(t *testing.T) {
	app.MigrateClean()
	defer app.DB.Close()

	req, _ := http.NewRequest("GET", "/api/projects", nil)
	req.Header.Set("Authorization", testUserToken1)
	response := executeRequest(req)

	assert.Equal(t, http.StatusOK, response.Code)

	body := response.Body.String()
	assert.Equal(t, "[]", body)
}

func TestProjectGetNonExistent(t *testing.T) {
	app.MigrateClean()
	defer app.DB.Close()

	req, _ := http.NewRequest("GET", "/api/project/"+uuid.NewV4().String(), nil)
	req.Header.Set("Authorization", testUserToken1)
	response := executeRequest(req)

	// TODO should be 404, need to fix it in middleware
	assert.Equal(t, http.StatusForbidden, response.Code)
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	assert.Equal(t, "forbidden", fmt.Sprintf("%s", m["error"]))
}

func TestProjectCreate(t *testing.T) {
	app.MigrateClean()
	defer app.DB.Close()

	var jsonStr = []byte(`{"name":"test project"}`)
	req, _ := http.NewRequest("POST", "/api/project", bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", testUserToken1)
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)
	assert.Equal(t, http.StatusCreated, response.Code)
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	name := fmt.Sprintf("%s", m["name"])
	assert.Equal(t, "test project", name)
	id := fmt.Sprintf("%s", m["id"])
	_, err := uuidParser.Parse(id)
	assert.Equal(t, nil, err)
}

func TestProjectGet(t *testing.T) {
	app.MigrateClean()
	defer app.DB.Close()

	var jsonStr = []byte(`{"name":"test project"}`)
	req, _ := http.NewRequest("POST", "/api/project", bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", testUserToken1)
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)
	assert.Equal(t, http.StatusCreated, response.Code)
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/project/%s", m["id"]), nil)
	req.Header.Set("Authorization", testUserToken1)
	response = executeRequest(req)
	assert.Equal(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/project/%s", m["id"]), nil)
	req.Header.Set("Authorization", testUserToken2)
	response = executeRequest(req)
	assert.Equal(t, http.StatusForbidden, response.Code)
}

func TestProjectUpdate(t *testing.T) {
	app.MigrateClean()
	defer app.DB.Close()

	// Add item
	var jsonStr = []byte(`{"name":"test project"}`)
	req, _ := http.NewRequest("POST", "/api/project", bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", testUserToken1)
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)
	assert.Equal(t, http.StatusCreated, response.Code)
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	name := fmt.Sprintf("%s", m["name"])
	assert.Equal(t, "test project", name)
	id := fmt.Sprintf("%s", m["id"])
	_, err := uuidParser.Parse(id)
	assert.Equal(t, nil, err)

	// Get the original
	req, _ = http.NewRequest("GET", "/api/project/"+id, nil)
	req.Header.Set("Authorization", testUserToken1)
	response = executeRequest(req)
	var originalProject map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalProject)

	// Update
	jsonStr = []byte(`{"name":"updated"}`)
	req, _ = http.NewRequest("PUT", "/api/project/"+id, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", testUserToken1)
	req.Header.Set("Content-Type", "application/json")
	response = executeRequest(req)
	assert.Equal(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/api/project/"+id, nil)
	req.Header.Set("Authorization", testUserToken1)
	response = executeRequest(req)
	var updated map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &updated)
	assert.Equal(t, originalProject["id"], updated["id"])
	assert.NotEqual(t, originalProject["name"], updated["name"])
	assert.Equal(t, "updated", fmt.Sprintf("%s", updated["name"]))
}

func TestProjectDelete(t *testing.T) {
	app.MigrateClean()
	defer app.DB.Close()

	// Add item
	var jsonStr = []byte(`{"name":"test project"}`)
	req, _ := http.NewRequest("POST", "/api/project", bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", testUserToken1)
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)
	assert.Equal(t, http.StatusCreated, response.Code)
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	name := fmt.Sprintf("%s", m["name"])
	assert.Equal(t, "test project", name)
	id := fmt.Sprintf("%s", m["id"])
	_, err := uuidParser.Parse(id)
	assert.Equal(t, nil, err)

	req, _ = http.NewRequest("GET", "/api/project/"+id, nil)
	req.Header.Set("Authorization", testUserToken1)
	response = executeRequest(req)
	assert.Equal(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/api/project/"+id, nil)
	req.Header.Set("Authorization", testUserToken1)
	response = executeRequest(req)
	assert.Equal(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/api/project/"+id, nil)
	req.Header.Set("Authorization", testUserToken1)
	response = executeRequest(req)
	assert.Equal(t, http.StatusNotFound, response.Code)
}
