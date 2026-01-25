package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"myapp/data"
	"net/http"
	"time"

	"github.com/CloudyKit/jet/v6"
	"github.com/ansufw/celeritas/mailer"
	"github.com/ansufw/celeritas/urlsigner"
)

func (h *Handlers) UserLogin(w http.ResponseWriter, r *http.Request) {

	err := h.App.Render.Page(w, r, "login", nil, nil)
	if err != nil {
		h.App.ErrorLog.Println(err)
		return
	}
}

func (h *Handlers) PostUserLogin(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		h.App.ErrorLog.Println(err)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := h.Models.Users.GetByEmail(email)
	if err != nil {
		h.App.ErrorLog.Println(err)
		return
	}

	matches, err := user.PasswordMatches(password)
	if err != nil {
		w.Write([]byte("error validating password"))
		return
	}

	if !matches {
		w.Write([]byte("Invalid credentials"))
		return
	}

	// did the user check the remember me box?
	if r.Form.Get("remember") == "remember" {
		randomString := h.randomString(20)
		hasher := sha256.New()
		_, err := hasher.Write([]byte(randomString))
		if err != nil {
			h.App.ErrorLog.Println(err)
			return
		}
		sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
		rm := data.RememberToken{}
		err = rm.InsertToken(user.ID, sha)
		if err != nil {
			h.App.ErrorStatus(w, http.StatusBadRequest)
			return
		}

		// set a cookie
		expire := time.Now().Add(365 * 24 * 60 * time.Second)
		cookie := &http.Cookie{
			Name:     h.App.AppName + "_remember_token",
			Value:    fmt.Sprintf("%d|%s", user.ID, sha),
			Path:     "/",
			Expires:  expire,
			HttpOnly: true,
			Domain:   h.App.Session.Cookie.Domain,
			MaxAge:   315_350_000,
			Secure:   h.App.Session.Cookie.Secure,
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, cookie)
		// save hash in session
		h.App.Session.Put(r.Context(), "remember_token", sha)
	}

	h.App.Session.Put(r.Context(), "user_id", user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {

	// delete remember token
	if h.App.Session.Exists(r.Context(), "remember_token") {
		rt := data.RememberToken{}
		err := rt.Delete(h.App.Session.GetString(r.Context(), "remember_token"))
		if err != nil {
			h.App.ErrorLog.Println(err)
		}
	}

	// delete cookie
	http.SetCookie(w, &http.Cookie{
		Name:     h.App.AppName + "_remember_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Domain:   h.App.Session.Cookie.Domain,
		MaxAge:   -1,
		Secure:   h.App.Session.Cookie.Secure,
		SameSite: http.SameSiteStrictMode,
	})

	h.App.Session.RenewToken(r.Context())
	h.App.Session.Remove(r.Context(), "userID")
	h.App.Session.Remove(r.Context(), "remember_token")
	h.App.Session.Destroy(r.Context())
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) Forgot(w http.ResponseWriter, r *http.Request) {
	err := h.render(w, r, "forgot", nil, nil)
	if err != nil {
		h.App.ErrorLog.Println("error rendering forgot page: ", err)
		h.App.Error500(w, r)
		return
	}
}

func (h *Handlers) PostForgot(w http.ResponseWriter, r *http.Request) {
	// parse form
	err := r.ParseForm()
	if err != nil {
		h.App.ErrorStatus(w, http.StatusBadRequest)
		return
	}

	// verify that supplied email exists
	var u *data.User
	email := r.Form.Get("email")
	log.Println(">> delme email: ", email)
	u, err = u.GetByEmail(email)
	if err != nil {
		h.App.ErrorLog.Println("error getting user by email: ", err)
		h.App.ErrorStatus(w, http.StatusBadRequest)
		return
	}

	// create a link to password reset form
	link := h.App.Server.URL + "/users/reset-password?email=" + email

	// sign the link
	sign := urlsigner.Signer{
		Secret: []byte(h.App.EncryptionKey),
	}
	signedLink := sign.GenerateTokenFromString(link)
	h.App.InfoLog.Println("Signed link is: ", signedLink)

	// email the message
	var data struct {
		Link string
	}
	data.Link = signedLink

	msg := mailer.Message{
		To:       u.Email,
		Subject:  "Reset your password",
		Template: "password-reset",
		Data:     data,
		From:     "admin@ansuf.my.id",
	}

	h.App.Mail.Jobs <- msg
	res := <-h.App.Mail.Results
	if res.Error != nil {
		h.App.ErrorLog.Println("error sending mail: ", res.Error)
		h.App.ErrorStatus(w, http.StatusBadRequest)
		return
	}

	// redirect the user
	http.Redirect(w, r, "/users/login", http.StatusSeeOther)
}

func (h *Handlers) ResetPasswordForm(w http.ResponseWriter, r *http.Request) {
	// get form values
	email := r.URL.Query().Get("email")
	theURL := r.RequestURI
	testURL := h.App.Server.URL + theURL

	// validate the url
	signer := urlsigner.Signer{
		Secret: []byte(h.App.EncryptionKey),
	}

	_, err := signer.VerifyToken(testURL)
	if err != nil {
		h.App.ErrorLog.Println("invalid token: ", err)
		h.App.ErrorUnauthorized(w, r)
		return
	}

	// make sure it's not expired
	expired := signer.Expired(testURL, 60)
	if expired {
		h.App.ErrorLog.Println("link expired")
		h.App.ErrorUnauthorized(w, r)
		return
	}

	// display form
	encryptedEmail, _ := h.encrypt(email)
	vars := make(jet.VarMap)
	vars.Set("email", encryptedEmail)

	err = h.render(w, r, "reset-password", vars, nil)
	if err != nil {
		h.App.ErrorLog.Println("error rendering reset password form: ", err)
		h.App.ErrorStatus(w, http.StatusBadRequest)
		return
	}

}

func (h *Handlers) PostResetPassword(w http.ResponseWriter, r *http.Request) {
	// parse form
	err := r.ParseForm()
	if err != nil {
		h.App.ErrorLog.Println("error parsing form: ", err)
		h.App.Error500(w, r)
		return
	}

	// get and decrypt email
	email, err := h.decrypt(r.Form.Get("email"))
	if err != nil {
		h.App.ErrorLog.Println("error decrypting email: ", err)
		h.App.Error500(w, r)
		return
	}

	// get the user
	var u *data.User
	u, err = u.GetByEmail(email)
	if err != nil {
		h.App.ErrorLog.Println("error getting user by email: ", err)
		h.App.Error500(w, r)
		return
	}

	// reset the password
	err = u.ResetPassword(u.ID, r.Form.Get("password"))
	if err != nil {
		h.App.ErrorLog.Println("error resetting password: ", err)
		h.App.Error500(w, r)
		return
	}

	// redirect the user
	h.App.Session.Put(r.Context(), "flash", "Password reset successfully")
	http.Redirect(w, r, "/users/login", http.StatusSeeOther)

}
