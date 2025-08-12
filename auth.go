package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

func (app *App) authenticateIDToken(idToken string) (*User, error) {
	var err error
	if len(idToken) < 1 {
		err = errors.New("invalid-token")
		log.Println(err)
		return nil, err
	}
	var u User
	hash := sha256.Sum256([]byte(idToken))
	t := Token{Key: hex.EncodeToString(hash[:])}
	if err = t.verifyCachedToken(); err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return nil, err
	}
	log.Println(t)

	if err == sql.ErrNoRows {
		err = nil

		var emailAddress string

		if flag.Lookup("test.v") != nil { // In test
			claims := jwt.MapClaims{}
			_, _ = jwt.ParseWithClaims(idToken, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(idToken), nil
			})
			if claims["email"] == nil {
				err = errors.New("invalid-token")
				log.Println(err)
				return nil, err
			}
			emailAddress = fmt.Sprintf("%v", claims["email"])
		} else {
			// Verify
			verified, err := app.Firebase.VerifyIDToken(context.Background(), idToken)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			emailAddress = fmt.Sprintf("%v", verified.Firebase.Identities["email"].([]interface{})[0])
		}

		if len(emailAddress) < 1 {
			log.Println(err)
			err = errors.New("invalid-token")
			return nil, err
		}

		// Create user
		u = User{EmailAddress: emailAddress}
		if err := u.createUser(); err != nil {
			log.Println(err)
			if strings.Contains(err.Error(), "duplicate") {
				if err = u.getUser(); err != nil {
					log.Println(err)
					return nil, err
				}
			} else {
				return nil, err
			}
		}

		// Ignoring err, should not be blocking
		emailAddresses := []string{}
		emailAddresses = append(emailAddresses, emailAddress)
		_ = sendEmailMessage(emailAddresses, "Thank you for signed up.", "Thank you for signed up at "+os.Getenv("APP_NAME"))

		// Store cached token
		t.EmailAddress = u.EmailAddress
		t.UserID = u.ID
		t.AuthProvider = "google"
		if err := t.cacheToken(); err != nil {
			log.Println(err)
			return nil, err
		}
	} else {
		u = User{EmailAddress: t.EmailAddress}
		if err = u.getUser(); err != nil {
			log.Println(err)
			return nil, err
		}
	}

	quotas, err := getUserQuotas(u.ID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	u.Quotas = quotas

	return &u, err
}
