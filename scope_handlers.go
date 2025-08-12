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

func (app *App) getScope(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["id"]
	_, err = uuidParser.Parse(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid-id")
		return
	}

	p := Scope{ID: id}
	if err = p.getScope(); err != nil {
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

func (app *App) getScopes(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))
	projectId := r.FormValue("projectId")

	if count > 10 || count < 1 {
		count = 1000
	}
	if start < 0 {
		start = 0
	}

	currentUser := r.Context().Value("currentUser")
	if currentUser == nil {
		log.Println("currentUser is empty")
		err := errors.New("current-user-is-empty")
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	access := Acl{
		ObjectID: projectId,
		UserID:   currentUser.(*User).ID,
	}
	err := access.getAccess()
	if err != nil {
		log.Println(err)
		if err == sql.ErrNoRows {
			respondError(w, http.StatusForbidden, "forbidden")
		} else {
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	scopes, err := getScopes(start, count, projectId)
	if err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	scenarios, err := getScenarios(start, 100000, projectId)
	if err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	for _, scen := range scenarios {
		index := -1
		for idx, scope := range scopes {
			if scope.ID == scen.ScopeID {
				index = idx
				break
			}
		}
		if index < 0 {
			p := Scope{ID: scen.ScopeID}
			err = p.getScope()
			if err != nil && err != sql.ErrNoRows {
				respondError(w, http.StatusInternalServerError, err.Error())
			}
			if err == nil {
				p.Scenarios = append(p.Scenarios, scen)
				scopes = append(scopes, p)
			}
		} else {
			scopes[index].Scenarios = append(scopes[index].Scenarios, scen)
		}
	}

	respond(w, http.StatusOK, scopes)
}

func (app *App) createScope(w http.ResponseWriter, r *http.Request) {
	var err error
	var p Scope
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
		isEligible, err := isEligibleToCreateScope(p.ProjectID)
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

	if err := p.createScope(); err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	currentUser := r.Context().Value("currentUser").(*User)
	access := Acl{
		ObjectID:   p.ID,
		ObjectType: "scope",
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

func (app *App) updateScope(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["id"]
	_, err = uuidParser.Parse(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid-id")
		return
	}

	var p Scope
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		log.Println(err)
		respondError(w, http.StatusBadRequest, "invalid-payload")
		return
	}
	defer r.Body.Close()
	p.ID = id

	if err := p.updateScope(); err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond(w, http.StatusOK, p)
}

func (app *App) deleteScope(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["id"]
	_, err = uuidParser.Parse(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid-id")
		return
	}

	p := Scope{ID: id}
	if err := p.deleteScope(); err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond(w, http.StatusOK, map[string]string{"result": "success"})
}
