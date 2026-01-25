package handlers

import (
	"fmt"
	"myapp/data"
	"net/http"
	"time"

	"github.com/CloudyKit/jet/v6"
	"github.com/ansufw/celeritas"
)

type XMLPayload struct {
	ID      int64    `xml:"id"`
	Name    string   `xml:"name"`
	Hobbies []string `xml:"hobby"`
}

type Handlers struct {
	App    *celeritas.Celeritas
	Models *data.Model
}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	defer h.App.LoadTime(time.Now())
	err := h.App.Render.Page(w, r, "home", nil, nil)
	if err != nil {
		h.App.ErrorLog.Println("Error rendering page: ", err)
	}
}

func (h *Handlers) JetPage(w http.ResponseWriter, r *http.Request) {
	err := h.App.Render.JetPage(w, r, "jet-template", nil, nil)
	if err != nil {
		h.App.ErrorLog.Println("Error rendering page: ", err)
	}
}

func (h *Handlers) GoPage(w http.ResponseWriter, r *http.Request) {
	err := h.App.Render.GoPage(w, r, "home", nil)
	if err != nil {
		h.App.ErrorLog.Println("Error rendering page: ", err)
	}
}

func (h *Handlers) SessionTest(w http.ResponseWriter, r *http.Request) {

	myData := "bar"

	h.App.Session.Put(r.Context(), "foo", myData)

	myValue := h.App.Session.Get(r.Context(), "foo")

	vars := make(jet.VarMap)
	vars.Set("foo", myValue)

	err := h.App.Render.JetPage(w, r, "sessions", vars, nil)
	if err != nil {
		h.App.ErrorLog.Println("Error rendering page: ", err)
	}
}

func (h *Handlers) JSON(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ID      int64    `json:"id"`
		Name    string   `json:"name"`
		Hobbies []string `json:"hobbies"`
	}

	payload.ID = 34
	payload.Name = "Jack Jones"
	payload.Hobbies = []string{"Reading", "Coding", "Hiking"}

	err := h.App.WriteJSON(w, http.StatusOK, payload)
	if err != nil {
		h.App.ErrorLog.Println("Error writing JSON: ", err)
		return
	}

}

func (h *Handlers) XML(w http.ResponseWriter, r *http.Request) {
	var payload XMLPayload

	payload.ID = 34
	payload.Name = "Jane Doe"
	payload.Hobbies = []string{"Reading", "Coding", "Hiking"}

	err := h.App.WriteXML(w, http.StatusOK, payload)
	if err != nil {
		h.App.ErrorLog.Println("Error writing XML: ", err)
		return
	}

}

func (h *Handlers) DownloadFile(w http.ResponseWriter, r *http.Request) {
	h.App.DownloadFile(w, r, "./public/images", "celeritas.jpg")
}

func (h *Handlers) TestCrypto(w http.ResponseWriter, r *http.Request) {
	plainText := "Hello World"
	fmt.Fprint(w, "Unencrypted: ", plainText+"\n")
	encryptedText, err := h.encrypt(plainText)
	if err != nil {
		h.App.ErrorLog.Println("Error encrypting text: ", err)
		return
	}
	fmt.Fprint(w, "Encrypted: ", encryptedText+"\n")
	decryptedText, err := h.decrypt(encryptedText)
	if err != nil {
		h.App.ErrorLog.Println("Error decrypting text: ", err)
		return
	}
	fmt.Fprint(w, "Decrypted: ", decryptedText+"\n")
}
