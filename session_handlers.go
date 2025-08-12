package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	uuidParser "github.com/docker/distribution/uuid"
	"github.com/gorilla/mux"
)

func (app *App) getSession(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["id"]
	_, err = uuidParser.Parse(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid-id")
		return
	}

	p := Session{ID: id}
	if err = p.getSession(); err != nil {
		log.Println(err)
		switch err {
		case sql.ErrNoRows:
			respondError(w, http.StatusNotFound, "item-not-found")
		default:
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	p.Scenarios, err = getScenariosBySession(0, 1000, p.ID)
	if err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	tests, err := getTests(0, 1000, p.ID)
	if err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for i, _ := range tests {
		for j, _ := range tests[i].Assists {
			u := User{}
			u.ID = tests[i].Assists[j].ID
			err := u.getUser()
			if err != nil {
				log.Println(err)
			}
			tests[i].Assists[j].Name = u.EmailAddress // TODO This should be the username
		}
	}

	for i, _ := range p.Scenarios {
		for j, _ := range tests {
			if p.Scenarios[i].ID == tests[j].ScenarioID {
				p.Scenarios[i].AssigneeID = tests[j].AssigneeID
				p.Scenarios[i].AssigneeName = tests[j].AssigneeName
				p.Scenarios[i].Status = tests[j].Status
				p.Scenarios[i].Assists = tests[j].Assists
				p.Scenarios[i].Notes = tests[j].Notes
			}
		}
	}

	respond(w, http.StatusOK, p)
}

func (app *App) getSessions(w http.ResponseWriter, r *http.Request) {
	log.Println(r.FormValue("count"))
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))
	projectId := r.FormValue("projectId")

	if count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	log.Println(count)
	log.Println(start)

	currentUser := r.Context().Value("currentUser")
	if currentUser == nil {
		log.Println("currentUser is empty")
		err := errors.New("current-user-is-empty")
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sessions, err := getSessions(start, count, projectId)
	if err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond(w, http.StatusOK, sessions)
}

func (app *App) createSession(w http.ResponseWriter, r *http.Request) {
	var err error
	var p Session
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		log.Println(err)
		respondError(w, http.StatusBadRequest, "invalid-payload")
		return
	}
	defer r.Body.Close()

	currentUser := r.Context().Value("currentUser").(*User)
	p.AuthorID = currentUser.ID

	gb := NewGrowthBook(app.GBFeatures, "")
	isContentCreationLimiterEnabled := gb.Feature(`content_creation_limiter`).On
	if isContentCreationLimiterEnabled {
		isEligible, err := isEligibleToCreateSession(p.ProjectID)
		if err != nil {
			log.Println(err)
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !isEligible {
			respondError(w, 429, "too-many-scopes")
			return
		}
	}

	if err = p.createSession(); err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	access := Acl{
		ObjectID:   p.ID,
		ObjectType: "session",
		UserID:     currentUser.ID,
		Access:     "OWNER",
	}
	err = access.createAccess()
	if err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond(w, http.StatusCreated, p)
}

func (app *App) updateSession(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["id"]
	_, err = uuidParser.Parse(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid-id")
		return
	}

	var p Session
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		log.Println(err)
		respondError(w, http.StatusBadRequest, "invalid-payload")
		return
	}
	defer r.Body.Close()
	p.ID = id

	if err := p.updateSession(); err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond(w, http.StatusOK, p)
}

func (app *App) deleteSession(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["id"]
	_, err = uuidParser.Parse(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid-id")
		return
	}

	p := Session{ID: id}
	if err := p.deleteSession(); err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond(w, http.StatusOK, map[string]string{"result": "success"})
}

func (app *App) resetSession(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["id"]
	_, err = uuidParser.Parse(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid-id")
		return
	}

	p := Session{ID: id}
	if err := p.resetSession(); err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond(w, http.StatusOK, map[string]string{"result": "success"})
}

func (app *App) createTest(w http.ResponseWriter, r *http.Request) {
	var p Test
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		log.Println(err)
		respondError(w, http.StatusBadRequest, "invalid-payload")
		return
	}
	defer r.Body.Close()

	currentUser := r.Context().Value("currentUser").(*User)
	/*
		0: unassigned
		1: ontest
		2: passed
		3: fail
	*/
	p.Status = 1
	p.AssigneeID = currentUser.ID

	// Check existing test by other
	err := p.getTestByOther()
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if p.ID != "" {
		respond(w, http.StatusConflict, p)
		return
	}

	// Check if this scenario has a test that is still on going
	err = p.getTestByAssignee()
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if p.ID != "" {
		for j, _ := range p.Assists {
			u := User{}
			u.ID = p.Assists[j].ID
			err := u.getUser()
			if err != nil {
				log.Println(err)
			}
			p.Assists[j].Name = u.EmailAddress // TODO This should be the username
		}
		respond(w, http.StatusCreated, p)
		return
	}

	// If there is no such test, create one
	// Get the scenario first
	s := Scenario{ID: p.ScenarioID}
	err = s.getScenario()
	if err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if s.ID != "" {
		p.Steps = s.Steps
	}
	if err := p.createTest(); err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	access := Acl{
		ObjectID:   p.ID,
		ObjectType: "test",
		UserID:     currentUser.ID,
		Access:     "OWNER",
	}
	err = access.createAccess()
	if err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond(w, http.StatusCreated, p)
}

func (app *App) deleteTest(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["id"]
	_, err = uuidParser.Parse(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid-id")
		return
	}

	p := Test{ID: id}
	if err := p.deleteTest(); err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond(w, http.StatusOK, map[string]string{"result": "success"})
}

func (app *App) updateTest(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["id"]
	_, err = uuidParser.Parse(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid-id")
		return
	}

	var p Test
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		log.Println(err)
		respondError(w, http.StatusBadRequest, "invalid-payload")
		return
	}
	defer r.Body.Close()
	p.ID = id

	if err := p.updateTest(); err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond(w, http.StatusOK, p)
}
