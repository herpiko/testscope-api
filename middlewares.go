package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

var PUBLIC_ENDPOINTS = [...]string{
	"/api/payments/callback", // Called by Xendit payment
	"/api/invite",
	"/static",
}

var PRIVATE_ENDPOINTS = [...]string{
	"/",
	"/api/invite",
}

// Generic middleware
func Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		isForbidden := false
		isUnauthorized := false
		// Authentication
		token := r.Header.Get("Authorization")
		if strings.Contains(token, "earer") {
			token = strings.Split(token, "earer ")[1]
		}
		currentUser, err := app.authenticateIDToken(token)
		if err != nil {
			log.Println(err)
			isUnauthorized = true
		}

		isPublic := false
		for _, endpoint := range PUBLIC_ENDPOINTS {
			if strings.HasPrefix(r.URL.Path, endpoint) {
				isPublic = true
				break
			}
		}
		if isPublic {
			log.Println("PUBLIC", r.Method, r.URL.Path) // Route log
			// We are not checking any credential
			// But if there is one, let pass it
			if currentUser != nil {
				log.Println("with user context")
				ctx := context.WithValue(r.Context(), "currentUser", currentUser)
				h.ServeHTTP(w, r.WithContext(ctx))
			} else {
				log.Println("without user context")
				h.ServeHTTP(w, r)
			}

			return
		}

		if isUnauthorized {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// Access control list by object id
		vars := mux.Vars(r)
		id := vars["id"]
		if currentUser == nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if currentUser.Role == "ADMIN" {
			// Admin can do anything, skip ACL
			log.Println(r.Method, r.URL.Path, "as", currentUser.EmailAddress, currentUser.Role) // Route log
		} else if currentUser.Role == "USER" && len(id) > 0 &&
			// Applied ACL
			(r.Method == "GET" ||
				r.Method == "PUT" ||
				r.Method == "DELETE") {

			access := Acl{
				ObjectID: id,
				UserID:   currentUser.ID,
			}
			err := access.getAccess()
			if err != nil {
				log.Println(err)
				if err == sql.ErrNoRows {
					log.Println("Forbidden 1")
					isForbidden = true
				} else {
					respondError(w, http.StatusInternalServerError, err.Error())
					return
				}
			}

			log.Println(r.Method, r.URL.Path, "as", currentUser.EmailAddress, currentUser.Role, access.Access) // Route log

			switch r.Method {
			case "GET":
				if !(access.Access == "READ" ||
					access.Access == "MODIFY" ||
					access.Access == "OWNER") {
					log.Println("Forbidden 2")
					isForbidden = true
				}
			case "PUT":
				if !(access.Access == "MODIFY" ||
					access.Access == "OWNER") {
					log.Println("Forbidden 3")
					isForbidden = true
				}
			case "DELETE":
				if !(access.Access == "OWNER") {
					log.Println("Forbidden 4")
					isForbidden = true
				}
			default:
				log.Println(r.Method)
			}
		} else {
			// Other cases those are not covered by ACL
			log.Println(r.Method, r.URL.Path, "as", currentUser.EmailAddress, currentUser.Role) // Route log
		}

		jsonBytes, _ := json.Marshal(currentUser)
		log.Println(string(jsonBytes))
		// Pass current user into the context
		ctx := context.WithValue(r.Context(), "currentUser", currentUser)

		ignore := false
		for _, endpoint := range PRIVATE_ENDPOINTS {
			if strings.Contains(r.URL.Path, endpoint) {
				ignore = true
				break
			}
		}

		if ignore {
			log.Println(r.Method, r.URL.Path) // Route log
			h.ServeHTTP(w, r.WithContext(ctx))
			return
		} else if isForbidden {
			log.Println("FORBIDDEN")
			respondError(w, http.StatusForbidden, "forbidden")
			return
		}

		/* Example on how to consume the context

		currentUser := r.Context().Value("currentUser")
		if currentUser == nil {
			log.Println("currentUser is empty")
		}
		log.Println(currentUser.(*user).EmailAddress)

		*/

		h.ServeHTTP(w, r.WithContext(ctx))

	})
}
