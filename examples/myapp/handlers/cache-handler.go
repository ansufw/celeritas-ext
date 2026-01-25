package handlers

import (
	"net/http"

	"github.com/justinas/nosurf"
)

func (h *Handlers) ShowCachePage(w http.ResponseWriter, r *http.Request) {
	err := h.render(w, r, "cache", nil, nil)
	if err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
		h.App.Error500(w, r)
	}
}

func (h *Handlers) SaveInCache(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		Name  string `json:"name"`
		Value string `json:"value"`
		CSRF  string `json:"csrf_token"`
	}

	err := h.App.ReadJSON(w, r, &userInput)
	if err != nil {
		h.App.ErrorLog.Println("error reading JSON:", err)
		h.App.Error500(w, r)
		return
	}

	if !nosurf.VerifyToken(nosurf.Token(r), userInput.CSRF) {
		h.App.Error500(w, r)
		return
	}

	err = h.App.Cache.Set(userInput.Name, userInput.Value)
	if err != nil {
		h.App.ErrorLog.Println("error saving in cache:", err)
		h.App.Error500(w, r)
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false
	resp.Message = "Saved in cache successfully"

	err = h.App.WriteJSON(w, http.StatusOK, resp)
	if err != nil {
		h.App.ErrorLog.Println("error writing JSON:", err)
		h.App.Error500(w, r)
		return
	}
}

func (h *Handlers) GetFromCache(w http.ResponseWriter, r *http.Request) {

	var msg string
	var inCache = true
	var userInput struct {
		Name string `json:"name"`
		CSRF string `json:"csrf_token"`
	}

	err := h.App.ReadJSON(w, r, &userInput)
	if err != nil {
		h.App.ErrorLog.Println("error reading JSON:", err)
		h.App.Error500(w, r)
		return
	}

	if !nosurf.VerifyToken(nosurf.Token(r), userInput.CSRF) {
		h.App.Error500(w, r)
		return
	}

	fromCache, err := h.App.Cache.Get(userInput.Name)
	if err != nil {
		msg = "not found in cache"
		inCache = false
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
		Value   string `json:"value"`
	}

	if inCache {
		resp.Error = false
		resp.Message = "Retrieved from cache successfully"
		resp.Value = fromCache.(string)
	} else {
		resp.Error = true
		resp.Message = msg
	}

	h.App.WriteJSON(w, http.StatusCreated, resp)
}

func (h *Handlers) DeleteFromCache(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		Name string `json:"name"`
		CSRF string `json:"csrf_token"`
	}

	err := h.App.ReadJSON(w, r, &userInput)
	if err != nil {
		h.App.ErrorLog.Println("error reading JSON:", err)
		h.App.Error500(w, r)
		return
	}

	if !nosurf.VerifyToken(nosurf.Token(r), userInput.CSRF) {
		h.App.Error500(w, r)
		return
	}

	err = h.App.Cache.Forget(userInput.Name)
	if err != nil {
		h.App.ErrorLog.Println("error deleting from cache:", err)
		h.App.Error500(w, r)
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
		Value   string `json:"value"`
	}

	resp.Error = false
	resp.Message = "Deleted from cache successfully"

	h.App.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handlers) EmptyCache(w http.ResponseWriter, r *http.Request) {

	var userInput struct {
		CSRF string `json:"csrf_token"`
	}

	err := h.App.ReadJSON(w, r, &userInput)
	if err != nil {
		h.App.ErrorLog.Println("error reading JSON:", err)
		h.App.Error500(w, r)
		return
	}

	if !nosurf.VerifyToken(nosurf.Token(r), userInput.CSRF) {
		h.App.Error500(w, r)
		return
	}

	err = h.App.Cache.Empty()
	if err != nil {
		h.App.ErrorLog.Println("error emptying cache:", err)
		h.App.Error500(w, r)
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false
	resp.Message = "emptied cache successfully"

	h.App.WriteJSON(w, http.StatusOK, resp)
}
