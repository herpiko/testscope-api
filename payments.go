package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"github.com/xendit/xendit-go"
	XenditInvoice "github.com/xendit/xendit-go/invoice"
)

func (app *App) createXenditInvoice(data *XenditInvoice.CreateParams) (*xendit.Invoice, error) {
	resp, err := app.Xendit.Invoice.Create(data)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return resp, nil
}

func (app *App) getXenditInvoice(data *XenditInvoice.GetParams) (*xendit.Invoice, error) {
	resp, err := app.Xendit.Invoice.Get(data)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return resp, nil
}

func (app *App) createInvoice(w http.ResponseWriter, r *http.Request) {
	var err error

	var i Invoice
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&i); err != nil {
		log.Println(err)
		respondError(w, http.StatusBadRequest, "invalid-payload")
		return
	}
	defer r.Body.Close()

	// Validate product and get the real amount
	// TODO iterate Items instead of picking only the first
	productItem := Product{ID: i.Items[0].ID}
	err = productItem.getProduct()
	if err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	i.Items[0].Amount = productItem.Amount

	currentUser := r.Context().Value("currentUser").(*User)

	xenditInvoice, err := app.createXenditInvoice(&XenditInvoice.CreateParams{
		ExternalID:  uuid.NewV4().String(),
		PayerEmail:  currentUser.EmailAddress,
		Description: productItem.Name,
		Amount:      productItem.Amount,
	})
	if err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	i.UserID = currentUser.ID
	i.URL = xenditInvoice.InvoiceURL
	i.Status = xenditInvoice.Status // PENDING
	i.ID = xenditInvoice.ExternalID
	i.ExternalID = xenditInvoice.ID
	i.EmailAddress = xenditInvoice.PayerEmail
	i.Amount = xenditInvoice.Amount
	i.Description = xenditInvoice.Description

	if err := i.createInvoice(); err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	access := Acl{
		ObjectID:   i.ID,
		ObjectType: "invoice",
		UserID:     currentUser.ID,
		Access:     "OWNER",
	}
	err = access.createAccess()
	if err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload := make(map[string]string)
	payload["external_id"] = xenditInvoice.ID
	payload["url"] = xenditInvoice.InvoiceURL

	respond(w, http.StatusCreated, payload)
}

func (app *App) paymentCallback(w http.ResponseWriter, r *http.Request) {
	var err error
	if r.Header.Get("x-callback-token") != os.Getenv("XENDIT_CALLBACK_TOKEN") {
		err = errors.New("invalid-callback-token")
		log.Println(err)
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	xenditInvoice := xendit.Invoice{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&xenditInvoice); err != nil {
		log.Println(err)
		respondError(w, http.StatusBadRequest, "invalid-payload")
		return
	}

	if xenditInvoice.Status == "PAID" || xenditInvoice.Status == "SETTLED" {
		i := Invoice{ExternalID: xenditInvoice.ID}
		i.Status = xenditInvoice.Status
		i.PaidAmount = xenditInvoice.PaidAmount
		i.PaymentMethod = xenditInvoice.PaymentMethod
		i.PaymentChannel = xenditInvoice.PaymentChannel
		i.PaymentDestination = xenditInvoice.PaymentDestination
		err = i.updateInvoice()
		if err != nil {
			log.Println(err)
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	respond(w, http.StatusOK, nil)
}

func (app *App) getInvoice(w http.ResponseWriter, r *http.Request) {
	var err error

	vars := mux.Vars(r)
	invoiceExternalID := vars["externalId"]

	xenditInvoice, err := app.getXenditInvoice(&XenditInvoice.GetParams{ID: invoiceExternalID})
	if err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if xenditInvoice.Status == "PAID" || xenditInvoice.Status == "SETTLED" {
		i := Invoice{ExternalID: xenditInvoice.ID}
		i.Status = xenditInvoice.Status
		i.PaidAmount = xenditInvoice.PaidAmount
		i.PaymentMethod = xenditInvoice.PaymentMethod
		i.PaymentChannel = xenditInvoice.PaymentChannel
		i.PaymentDestination = xenditInvoice.PaymentDestination
		err = i.updateInvoice()
		if err != nil {
			log.Println(err)
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	payload := make(map[string]string)
	payload["external_id"] = xenditInvoice.ID
	payload["url"] = xenditInvoice.InvoiceURL
	payload["status"] = xenditInvoice.Status
	respond(w, http.StatusOK, payload)
}

func (app *App) getInvoiceByUserID(w http.ResponseWriter, r *http.Request) {
	var err error

	vars := mux.Vars(r)
	userID := vars["userId"]

	i := Invoice{UserID: userID}
	err = i.getInvoice()
	if err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	xenditInvoice, err := app.getXenditInvoice(&XenditInvoice.GetParams{ID: i.ExternalID})
	if err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if xenditInvoice.Status == "PAID" || xenditInvoice.Status == "SETTLED" {
		i := Invoice{ExternalID: xenditInvoice.ID}
		i.Status = xenditInvoice.Status
		i.PaidAmount = xenditInvoice.PaidAmount
		i.PaymentMethod = xenditInvoice.PaymentMethod
		i.PaymentChannel = xenditInvoice.PaymentChannel
		i.PaymentDestination = xenditInvoice.PaymentDestination
		err = i.updateInvoice()
		if err != nil {
			log.Println(err)
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	payload := make(map[string]string)
	payload["external_id"] = xenditInvoice.ID
	payload["url"] = xenditInvoice.InvoiceURL
	payload["status"] = xenditInvoice.Status
	respond(w, http.StatusOK, payload)
}
