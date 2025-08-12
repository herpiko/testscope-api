package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	"github.com/growthbook/growthbook-golang"
	_ "github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/posthog/posthog-go"
	"github.com/xendit/xendit-go/client"
	"google.golang.org/api/option"
)

type App struct {
	Router     *mux.Router
	DB         *sql.DB
	Firebase   *auth.Client
	Xendit     *client.API
	Storage    *Minio
	GBFeatures growthbook.FeatureMap
	Posthog    posthog.Client
}

func (app *App) Init() {
	var err error

	// Regular migration
	err = app.migrateUp()
	if err != nil {
		log.Println(err)
		log.Fatal(err)
	}

	// Main database connection
	connectionString :=
		fmt.Sprintf("host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASS"),
			os.Getenv("DB_NAME"),
		)
	app.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Println(err)
		log.Fatal(err)
	}

	// File storage
	s3Endpoint := os.Getenv("S3_URL")
	s3AccessKey := os.Getenv("S3_ACCESS_KEY")
	s3SecretKey := os.Getenv("S3_SECRET_KEY")
	useSSL := false

	app.Storage = &Minio{
		bucketName:     "default-default",
		bucketLocation: "us-east-1",
	}

	cred := credentials.NewStaticV4(s3AccessKey, s3SecretKey, "")
	// Initialize minio client object.
	app.Storage.client, err = minio.New(s3Endpoint, &minio.Options{
		Creds:  cred,
		Secure: useSSL,
	})
	if err != nil {
		log.Println(s3Endpoint)
		log.Println(err)
		log.Fatal(err)
	}

	ctx, delay := context.WithTimeout(context.Background(), 5*time.Second)
	defer delay()
	err = app.Storage.client.MakeBucket(ctx, app.Storage.bucketName, minio.MakeBucketOptions{
		Region: app.Storage.bucketLocation,
	})
	// Prepare other buckets
	err = app.Storage.client.MakeBucket(ctx, PUBLIC_BUCKET, minio.MakeBucketOptions{
		Region: app.Storage.bucketLocation,
	})
	if err != nil {
		exists, errBucketExists := app.Storage.client.BucketExists(ctx, PUBLIC_BUCKET)
		if errBucketExists == nil && exists {
			// ignore
		} else {
			log.Println(errBucketExists)
			log.Fatal(errBucketExists)
		}
	}

	// Firebase connection
	opt := option.WithCredentialsFile(os.Getenv("FIREBASE_ACCOUNT_KEY_PATH"))
	config := &firebase.Config{ProjectID: os.Getenv("FIREBASE_PROJECT_ID")}

	firebaseApp, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		log.Println(err)
		log.Fatal(err)
	}
	app.Firebase, err = firebaseApp.Auth(context.Background())
	if err != nil {
		log.Println(err)
		log.Fatal(err)
	}

	// Xendit payment
	app.Xendit = client.New(os.Getenv("XENDIT_API_SECRET"))

	go func() {
		for {
			// Update only if the hash is different
			features, _ := GetGrowthbookFeatures()
			hash, _ := hashstructure.Hash(features, hashstructure.FormatV2, nil)
			existingHash, _ := hashstructure.Hash(app.GBFeatures, hashstructure.FormatV2, nil)
			if hash != existingHash {
				app.GBFeatures = features
			}
			time.Sleep(5 * time.Second)
		}
	}()

	// Posthog
	/*
		posthogApiKey := os.Getenv("POSTHOG_API_KEY")
		posthogPersonalApiKey := os.Getenv("POSTHOG_PERSONAL_API_KEY")
		app.Posthog, err = posthog.NewWithConfig(
			posthogApiKey,
			posthog.Config{
				//Endpoint:       "https://app.posthog.com/api",
				PersonalApiKey: posthogPersonalApiKey,
			},
		)
		if err != nil {
			panic(err)
		}
		// Reloading posthog
		log.Println("Reloading posthog ff")
		app.Posthog.ReloadFeatureFlags()
		log.Println("FF Reloaded")
	*/

	app.Router = mux.NewRouter()

	// Midlewares
	app.Router.Use(Middleware)

	app.initRoutes()
}

func (app *App) migrateUp() error {
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
		return err
	}

	if len(os.Getenv("MIGRATE_FORCE")) > 0 {
		// Force specific migration version
		migrateVersion, err := strconv.Atoi(os.Getenv("MIGRATE_FORCE"))
		if err != nil {
			log.Println(err)
			return err
		}
		err = migration.Force(migrateVersion)
		if err != nil && err.Error() != "no change" {
			log.Println(err)
			return err
		}
	} else {
		// Regular migration
		err = migration.Up()
		if err != nil && err.Error() != "no change" {
			log.Println(err)
			return err
		}
	}

	_, err = migration.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (app *App) initRoutes() {

	// Projects
	app.Router.HandleFunc("/api/projects", app.getProjects).Methods("GET")
	app.Router.HandleFunc("/api/project", app.createProject).Methods("POST")
	app.Router.HandleFunc("/api/project/{id}", app.getProject).Methods("GET")
	app.Router.HandleFunc("/api/project/{id}", app.updateProject).Methods("PUT")
	app.Router.HandleFunc("/api/project/{id}", app.deleteProject).Methods("DELETE")
	app.Router.HandleFunc("/api/invite/{id}", app.getInvitation).Methods("GET")
	app.Router.HandleFunc("/api/invite/{id}", app.acceptInvitation).Methods("PUT")
	app.Router.HandleFunc("/api/collaborators/{id}", app.getCollaborators).Methods("GET")
	app.Router.HandleFunc("/api/revoke/{projectId}/{userId}", app.revokeCollaborator).Methods("PUT")

	// Scopes
	app.Router.HandleFunc("/api/scopes", app.getScopes).Methods("GET")
	app.Router.HandleFunc("/api/scope", app.createScope).Methods("POST")
	app.Router.HandleFunc("/api/scope/{id}", app.getScope).Methods("GET")
	app.Router.HandleFunc("/api/scope/{id}", app.updateScope).Methods("PUT")
	app.Router.HandleFunc("/api/scope/{id}", app.deleteScope).Methods("DELETE")

	// Scenarios
	app.Router.HandleFunc("/api/scenarios", app.getScenarios).Methods("GET")
	app.Router.HandleFunc("/api/scenario", app.createScenario).Methods("POST")
	app.Router.HandleFunc("/api/scenario/{id}", app.getScenario).Methods("GET")
	app.Router.HandleFunc("/api/scenario/{id}", app.updateScenario).Methods("PUT")
	app.Router.HandleFunc("/api/scenario/{id}", app.deleteScenario).Methods("DELETE")

	// Sessions
	app.Router.HandleFunc("/api/sessions", app.getSessions).Methods("GET")
	app.Router.HandleFunc("/api/session", app.createSession).Methods("POST")
	app.Router.HandleFunc("/api/session/{id}", app.getSession).Methods("GET")
	app.Router.HandleFunc("/api/session/{id}", app.updateSession).Methods("PUT")
	app.Router.HandleFunc("/api/session/{id}", app.deleteSession).Methods("DELETE")
	app.Router.HandleFunc("/api/reset-session/{id}", app.resetSession).Methods("PUT")
	app.Router.HandleFunc("/api/test", app.createTest).Methods("POST")
	app.Router.HandleFunc("/api/test/{id}", app.deleteTest).Methods("DELETE")
	app.Router.HandleFunc("/api/test/{id}", app.updateTest).Methods("PUT")

	// Users
	app.Router.HandleFunc("/api/users", app.getUsers).Methods("GET")
	app.Router.HandleFunc("/api/user", app.createUser).Methods("POST")
	app.Router.HandleFunc("/api/user", app.getUser).Methods("GET")      // Without ID,
	app.Router.HandleFunc("/api/user/{id}", app.getUser).Methods("GET") // With ID
	app.Router.HandleFunc("/api/user/{id}", app.updateUser).Methods("PUT")
	app.Router.HandleFunc("/api/user/{id}", app.deleteUser).Methods("DELETE")

	// Payment
	app.Router.HandleFunc("/api/payments/callback", app.paymentCallback).Methods("POST")
	app.Router.HandleFunc("/api/payments/invoice", app.createInvoice).Methods("POST")
	app.Router.HandleFunc("/api/payments/invoice/{externalId}", app.getInvoice).Methods("GET")
	app.Router.HandleFunc("/api/payments/invoice-by-user-id/{userId}", app.getInvoiceByUserID).Methods("GET")

	// Blob
	app.Router.HandleFunc("/api/blob", app.uploadFile).Methods("POST")
	app.Router.HandleFunc("/api/blob/{id}", app.getFile).Methods("GET")
}

func (app *App) Run(addr string) {
	log.Println("Running on port ", addr)
	log.Fatal(http.ListenAndServe(addr, app.Router))
}
