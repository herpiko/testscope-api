package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	uuidParser "github.com/docker/distribution/uuid"
	"github.com/gorilla/mux"
)

func (app *App) getUser(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["id"]
	if len(id) > 0 {
		log.Println(id)
		_, err = uuidParser.Parse(id)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid-id")
			return
		}

		u := User{}
		u.ID = id
		if err = u.getUser(); err != nil {
			switch err {
			case sql.ErrNoRows:
				respondError(w, http.StatusNotFound, "item-not-found")
			default:
				respondError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
		respond(w, http.StatusOK, u)

	} else {
		if r.Context().Value("currentUser") != nil {
			respond(w, http.StatusOK, r.Context().Value("currentUser").(*User))
		} else {
			token := r.Header.Get("Authorization")
			if strings.Contains(token, "earer") {
				token = strings.Split(token, "earer ")[1]
			}
			log.Println(token)
			currentUser, err := app.authenticateIDToken(token)
			if err != nil {
				respondError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			respond(w, http.StatusOK, currentUser)
		}
	}
}

func (app *App) getUsers(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	users, err := getUsers(start, count)
	if err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond(w, http.StatusOK, users)
}

func (app *App) createUser(w http.ResponseWriter, r *http.Request) {
	var p User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		log.Println(err)
		respondError(w, http.StatusBadRequest, "invalid-payload")
		return
	}
	defer r.Body.Close()

	if err := p.createUser(); err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// This user have owner access to itself.
	access := Acl{
		ObjectID:   p.ID,
		ObjectType: "user",
		UserID:     p.ID,
		Access:     "OWNER",
	}
	err := access.createAccess()
	if err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// If it created by other user, let's add it as OWNER as well.
	if p.ID != r.Context().Value("currentUser").(*User).ID {
		access = Acl{
			ObjectID:   p.ID,
			ObjectType: "user",
			UserID:     r.Context().Value("currentUser").(*User).ID,
			Access:     "OWNER",
		}
		err := access.createAccess()
		if err != nil {
			log.Println(err)
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	respond(w, http.StatusCreated, p)
}

func (app *App) updateUser(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["id"]
	_, err = uuidParser.Parse(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid-id")
		return
	}

	var p User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		log.Println(err)
		respondError(w, http.StatusBadRequest, "invalid-payload")
		return
	}
	defer r.Body.Close()
	p.ID = id

	if err := p.updateUser(); err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond(w, http.StatusOK, p)
}

func (app *App) deleteUser(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["id"]
	_, err = uuidParser.Parse(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid-id")
		return
	}

	p := User{ID: id}
	if err := p.deleteUser(); err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond(w, http.StatusOK, map[string]string{"result": "success"})
}
