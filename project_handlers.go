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

func (app *App) getProject(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["id"]
	_, err = uuidParser.Parse(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid-id")
		return
	}

	p := Project{ID: id}
	if err = p.getProject(); err != nil {
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

func (app *App) getProjects(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
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

	projects, err := getProjects(start, count, currentUser.(*User).ID)
	if err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond(w, http.StatusOK, projects)
}

func (app *App) createProject(w http.ResponseWriter, r *http.Request) {
	var err error
	var p Project
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		log.Println(err)
		respondError(w, http.StatusBadRequest, "invalid-payload")
		return
	}
	defer r.Body.Close()

	currentUser := r.Context().Value("currentUser").(*User)

	gb := NewGrowthBook(app.GBFeatures, "")
	isContentCreationLimiterEnabled := gb.Feature(`content_creation_limiter`).On
	if isContentCreationLimiterEnabled {
		isEligible, err := isEligibleToCreateProject(currentUser.ID)
		if err != nil {
			log.Println(err)
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !isEligible {
			respondError(w, 429, "too-many-projects")
			return
		}
	}

	if err = p.createProject(); err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	access := Acl{
		ObjectID:   p.ID,
		ObjectType: "project",
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

func (app *App) updateProject(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["id"]
	_, err = uuidParser.Parse(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid-id")
		return
	}

	var p Project
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		log.Println(err)
		respondError(w, http.StatusBadRequest, "invalid-payload")
		return
	}
	defer r.Body.Close()
	p.ID = id

	if err := p.updateProject(); err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond(w, http.StatusOK, p)
}

func (app *App) deleteProject(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["id"]
	_, err = uuidParser.Parse(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid-id")
		return
	}

	p := Project{ID: id}
	if err := p.deleteProject(); err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond(w, http.StatusOK, map[string]string{"result": "success"})
}

func (app *App) getInvitation(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["id"]
	_, err = uuidParser.Parse(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid-id")
		return
	}

	p := Project{InviteCode: id}
	if err = p.getInvitation(); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondError(w, http.StatusNotFound, "item-not-found")
		default:
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	if r.Context().Value("currentUser") != nil {
		currentUser := r.Context().Value("currentUser").(*User)
		access := Acl{
			ObjectID: p.ID,
			UserID:   currentUser.ID,
		}
		err = access.getAccess()
		if err != nil && err != sql.ErrNoRows {
			log.Println(err)
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		p.Access = access.Access
	}

	respond(w, http.StatusOK, p)
}

func (app *App) acceptInvitation(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["id"]
	_, err = uuidParser.Parse(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid-id")
		return
	}

	p := Project{InviteCode: id}
	if err = p.getInvitation(); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondError(w, http.StatusNotFound, "item-not-found")
		default:
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	currentUser := r.Context().Value("currentUser").(*User)
	access := Acl{
		ObjectID:   p.ID,
		ObjectType: "project",
		UserID:     currentUser.ID,
		Access:     "MODIFY",
	}
	err = access.createAccess()
	if err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond(w, http.StatusOK, p)
}

func (app *App) getCollaborators(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["id"]
	_, err = uuidParser.Parse(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid-id")
		return
	}

	collaborators, err := getCollaborators(id)
	if err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond(w, http.StatusOK, collaborators)
}

func (app *App) revokeCollaborator(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	projectId := vars["projectId"]
	_, err = uuidParser.Parse(projectId)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid-id")
		return
	}

	userId := vars["userId"]
	_, err = uuidParser.Parse(userId)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid-id")
		return
	}

	access := Acl{
		ObjectID:   projectId,
		ObjectType: "project",
		UserID:     userId,
	}
	err = access.dropAccess()
	if err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond(w, http.StatusOK, nil)
}
