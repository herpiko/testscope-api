package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

var testAdminToken1, testUserToken1, testUserToken2 string

func TestMain(m *testing.M) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	_ = godotenv.Load()
	app = App{}
	app.MigrateInit()
	app.Init()

	// herpiko@gmail.com
	testAdminToken1 = "eyJhbGciOiJSUzI1NiIsImtpZCI6IjkwMDk1YmM2ZGM2ZDY3NzkxZDdkYTFlZWIxYTU1OWEzZDViMmM0ODYiLCJ0eXAiOiJKV1QifQ.eyJuYW1lIjoiSGVycGlrbyBEd2kgQWd1bm8iLCJwaWN0dXJlIjoiaHR0cHM6Ly9saDMuZ29vZ2xldXNlcmNvbnRlbnQuY29tL2EtL0FPaDE0R2pOeWROSDgxUm1Lczk4LWwtVVg1d2VTVUJZM0p4YTNiQVNZU0w4enc9czk2LWMiLCJpc3MiOiJodHRwczovL3NlY3VyZXRva2VuLmdvb2dsZS5jb20vdGVzdHNjb3BlLWlvIiwiYXVkIjoidGVzdHNjb3BlLWlvIiwiYXV0aF90aW1lIjoxNjI5OTk1NTc3LCJ1c2VyX2lkIjoiM01sbGY3b0xZM2czekZYNGNFbWJyOUtlVzk3MiIsInN1YiI6IjNNbGxmN29MWTNnM3pGWDRjRW1icjlLZVc5NzIiLCJpYXQiOjE2Mjk5OTU1NzgsImV4cCI6MTYyOTk5OTE3OCwiZW1haWwiOiJoZXJwaWtvQGdtYWlsLmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJmaXJlYmFzZSI6eyJpZGVudGl0aWVzIjp7Imdvb2dsZS5jb20iOlsiMTE2ODM4MDc1NDA2MDU4Nzc2NTI4Il0sImVtYWlsIjpbImhlcnBpa29AZ21haWwuY29tIl19LCJzaWduX2luX3Byb3ZpZGVyIjoiZ29vZ2xlLmNvbSJ9fQ.aVmu9xO8ss88ZtaWJqfxVGkQTWS-Q_u8Vo8gVIunrSe7dKQOnLaXz7JJLY8M4OBxVaBNZ9kLi5sncxTwWL38IJ6J5f85EIFcnWFS9nozJ8givk8S6vYDkOzfkMZpYZItoT6qf5n3rSHga-AASUrik_CaPCnG1ZUUU6ejpzscxzPejJogi7KrHTDxZmUxfOAnFsA7CqLYFaO641sR8WkkXHn4EdqLkzGg1R9N6ypDXKyrNnyoTCBeeIJbvNS8iJF0G6Yt4Dr3q6QcEALpjatD5QqwQopGp_3luubsHfAGr0t3kZ3SwfHz4Lb2evoEYuWlAhBxFPXeQ227x_o155xNbw"

	// pdafoot.tgz@gmail.com
	testUserToken1 = "eyJhbGciOiJSUzI1NiIsImN0eSI6IkpXVCJ9.eyJuYW1lIjoiRGFtYXIgR3VtaWxhbmciLCJwaWN0dXJlIjoiaHR0cHM6Ly9saDMuZ29vZ2xldXNlcmNvbnRlbnQuY29tL2EtL0FPaDE0R2pOeWROSDgxUm1Lczk4LWwtVVg1d2VTVUJZM0p4YTNiQVNZU0w4enc9czk2LWMiLCJpc3MiOiJodHRwczovL3NlY3VyZXRva2VuLmdvb2dsZS5jb20vdGVzdHNjb3BlLWlvIiwiYXVkIjoidGVzdHNjb3BlLWlvIiwiYXV0aF90aW1lIjoxNjI5OTk1NTc3LCJ1c2VyX2lkIjoiM01sbGY3b0xZM2czekZYNGNFbWJyOUtlVzk3MiIsInN1YiI6IjNNbGxmN29MWTNnM3pGWDRjRW1icjlLZVc5NzIiLCJpYXQiOjE2Mjk5OTU1NzgsImV4cCI6MTYyOTk5OTE3OCwiZW1haWwiOiJwYWRmb290LnRnekBnbWFpbC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZmlyZWJhc2UiOnsiaWRlbnRpdGllcyI6eyJnb29nbGUuY29tIjpbIjExNjgzODA3NTQwNjA1ODc3NjUyOSJdLCJlbWFpbCI6WyJwYWRmb290LnRnekBnbWFpbC5jb20iXX0sInNpZ25faW5fcHJvdmlkZXIiOiJnb29nbGUuY29tIn19.RAJjkJlD5Toq0sNDz2iYCmJu2rSsnavTyWw9rn6jKGW7yZ8mbNItyZByuFNoSU5CXqUD7Hz2iygYC6PeU7xHu0jXO8NVK5G1mWPBBjtGbdNXE5Zcn9ZpviBZ4drzMF6WLz33xoUD9Ce_0jvCQgLM_TvRQLap4dqD1uPCo7Hmy9M"

	// masepindrayana@gmail.com
	testUserToken2 = "eyJhbGciOiJSUzI1NiIsImN0eSI6IkpXVCJ9.eyJuYW1lIjoiTSBBc2VwIEluZHJheWFuYSIsInBpY3R1cmUiOiJodHRwczovL2xoMy5nb29nbGV1c2VyY29udGVudC5jb20vYS0vQU9oMTRHak55ZE5IODFSbUtzOTgtbC1VWDV3ZVNVQlkzSnhhM2JBU1lTTDh6dz1zOTYtYyIsImlzcyI6Imh0dHBzOi8vc2VjdXJldG9rZW4uZ29vZ2xlLmNvbS90ZXN0c2NvcGUtaW8iLCJhdWQiOiJ0ZXN0c2NvcGUtaW8iLCJhdXRoX3RpbWUiOjE2Mjk5OTU1NzcsInVzZXJfaWQiOiIzTWxsZjdvTFkzZzN6Rlg0Y0VtYnI5S2VXOTcyIiwic3ViIjoiM01sbGY3b0xZM2czekZYNGNFbWJyOUtlVzk3MiIsImlhdCI6MTYyOTk5NTU3OCwiZXhwIjoxNjI5OTk5MTc4LCJlbWFpbCI6Im1hc2VwaW5kcmF5YW5hQGdtYWlsLmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJmaXJlYmFzZSI6eyJpZGVudGl0aWVzIjp7Imdvb2dsZS5jb20iOlsiMTE2ODM4MDc1NDA2MDU4Nzc2NTMwIl0sImVtYWlsIjpbIm1hc2VwaW5kcmF5YW5hQGdtYWlsLmNvbSJdfSwic2lnbl9pbl9wcm92aWRlciI6Imdvb2dsZS5jb20ifX0.CFGy_mgP3nKfxZ0YsAOdRW-u6_eq0s9cUEyHLUtsvt-5axOLINkbXOjyLzb4fo1nzDsos_okpEmjZIOhZGrvBdbuYQSNz3ydKGuywJgO8AOwLk3WSTWYco2zHqJX4__hTMNkS3DhqDOXMRCf-2wF2fCmQ-6-KC3h5UEMdq8ajXM"
	code := m.Run()
	os.Exit(code)
}

func (a *App) MigrateInit() error {
	cwd, _ := os.Getwd()
	migrationPath := "file://" + cwd + "/migrations"
	connectionString :=
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASS"),
			os.Getenv("DB_HOST"),
			"5432",
			os.Getenv("DB_NAME"),
		)

	var err error

	migration, err := migrate.New(
		migrationPath,
		connectionString)
	if err != nil {
		log.Println(err)
	}
	err = migration.Drop()
	if err != nil && err.Error() != "no change" {
		log.Println(err)
	}
	_, _ = migration.Close()
	migration, err = migrate.New(
		migrationPath,
		connectionString)
	if err != nil {
		log.Println(err)
	}
	err = migration.Up()
	if err != nil && err.Error() != "no change" {
		log.Println(err)
	}
	_, _ = migration.Close()

	// Backup migrated db
	connectionString =
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASS"),
			os.Getenv("DB_HOST"),
			"5432",
			"postgres",
		)
	app.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Println(err)
		return err
	}
	_, _ = app.DB.Exec(`
	DROP DATABASE ` + os.Getenv("DB_NAME") + `_ready`)
	_, err = app.DB.Exec(`
	CREATE DATABASE ` + os.Getenv("DB_NAME") + `_ready WITH TEMPLATE ` + os.Getenv("DB_NAME"))
	if err != nil {
		log.Fatal(err)
		return err
	}

	app.DB.Close()
	return nil
}

func (a *App) MigrateClean() error {
	app.DB.Close()
	connectionString :=
		fmt.Sprintf("host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASS"),
			"postgres",
		)
	var err error

	app.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = app.DB.Exec("DROP DATABASE " + os.Getenv("DB_NAME"))
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = app.DB.Exec("CREATE DATABASE " + os.Getenv("DB_NAME") + " WITH TEMPLATE " + os.Getenv("DB_NAME") + "_ready")
	if err != nil {
		log.Println(err)
		return err
	}

	// Reconnect again using testdb
	app.DB.Close()
	connectionString =
		fmt.Sprintf("host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASS"),
			"testdb",
		)

	app.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)

	return rr
}
