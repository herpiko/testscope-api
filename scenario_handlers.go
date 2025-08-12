package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	uuidParser "github.com/docker/distribution/uuid"
	"github.com/gorilla/mux"
)

func (app *App) getScenario(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["id"]
	_, err = uuidParser.Parse(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid-id")
		return
	}

	p := Scenario{ID: id}
	if err = p.getScenario(); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondError(w, http.StatusNotFound, "item-not-found")
		default:
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respond(w, http.StatusOK, p)
}

func (app *App) getScenarios(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))
	projectId := r.FormValue("projectId")

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	scenarios, err := getScenarios(start, count, projectId)
	if err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, scenarios)
}

func (app *App) createScenario(w http.ResponseWriter, r *http.Request) {
	var err error
	var p Scenario
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		log.Println(err)
		respondError(w, http.StatusBadRequest, "invalid-payload")
		return
	}
	defer r.Body.Close()

	gb := NewGrowthBook(app.GBFeatures, "")
	isContentCreationLimiterEnabled := gb.Feature(`content_creation_limiter`).On
	if isContentCreationLimiterEnabled {
		isEligible, err := isEligibleToCreateScenario(p.ProjectID)
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

	if err := p.createScenario(); err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	currentUser := r.Context().Value("currentUser").(*User)
	access := Acl{
		ObjectID:   p.ID,
		ObjectType: "scenario",
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

func (app *App) updateScenario(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["id"]
	_, err = uuidParser.Parse(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid-id")
		return
	}

	var p Scenario
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		log.Println(err)
		respondError(w, http.StatusBadRequest, "invalid-payload")
		return
	}
	defer r.Body.Close()
	p.ID = id

	if err := p.updateScenario(); err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond(w, http.StatusOK, p)
}

func (app *App) deleteScenario(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["id"]
	_, err = uuidParser.Parse(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid-id")
		return
	}

	p := Scenario{ID: id}
	if err := p.deleteScenario(); err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond(w, http.StatusOK, map[string]string{"result": "success"})
}
