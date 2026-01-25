package main

import (
	"fmt"
	"myapp/data"
	"myapp/public"
	"net/http"
	"strconv"

	"github.com/ansufw/celeritas/mailer"
	"github.com/go-chi/chi/v5"
)

func (a *application) routes() *chi.Mux {
	a.use(a.Middleware.CheckRemember)

	a.get("/", a.handlers.Home)
	a.get("/go-page", a.handlers.GoPage)
	a.get("/jet-page", a.handlers.JetPage)
	a.get("/session", a.handlers.SessionTest)

	a.get("/users/login", a.handlers.UserLogin)
	a.post("/users/login", a.handlers.PostUserLogin)
	a.get("/users/logout", a.handlers.Logout)
	a.get("/users/forgot-password", a.handlers.Forgot)
	a.post("/users/forgot-password", a.handlers.PostForgot)
	a.get("/users/reset-password", a.handlers.ResetPasswordForm)
	a.post("/users/reset-password", a.handlers.PostResetPassword)

	a.get("/form", a.handlers.Form)
	a.post("/form", a.handlers.PostForm)

	a.get("/json", a.handlers.JSON)
	a.get("/xml", a.handlers.XML)
	a.get("/download", a.handlers.DownloadFile)

	a.get("/crypto", a.handlers.TestCrypto)
	a.get("/cache", a.handlers.ShowCachePage)
	a.post("/api/save-in-cache", a.handlers.SaveInCache)
	a.post("/api/get-from-cache", a.handlers.GetFromCache)
	a.post("/api/delete-from-cache", a.handlers.DeleteFromCache)
	a.post("/api/empty-cache", a.handlers.EmptyCache)

	a.get("/test-mail", func(w http.ResponseWriter, r *http.Request) {
		msg := mailer.Message{
			From:        "test@example.com",
			To:          "ansufw@gmail.com",
			Subject:     "Test Mail subject",
			Template:    "test",
			Attachments: nil,
			Data:        nil,
		}

		// via channel
		// a.App.Mail.Jobs <- msg
		// res := <-a.App.Mail.Results
		// if res.Error != nil {
		// 	a.App.ErrorLog.Println(res.Error)
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	return
		// }

		// direct send
		err := a.App.Mail.SendSMTPMessage(msg)
		if err != nil {
			a.App.ErrorLog.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		a.App.InfoLog.Println("Mail sent successfully")
	})

	a.get("/create-user", func(w http.ResponseWriter, r *http.Request) {
		u := data.User{
			FirstName:  "Rahmat",
			LastName:   "Wahab",
			Email:      "rahmat.wahab@example.com",
			Password:   "password",
			UserActive: 1,
		}

		id, err := a.models.Users.Insert(u)
		if err != nil {
			a.App.ErrorLog.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "User created with ID: %d", id)
	})

	a.get("/get-all-users", func(w http.ResponseWriter, r *http.Request) {
		users, err := a.models.Users.GetAll()
		if err != nil {
			a.App.ErrorLog.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		for _, user := range users {
			fmt.Fprintf(w, "User ID: %d | Name: %s %s | Email: %s\n", user.ID, user.FirstName, user.LastName, user.Email)
		}
	})

	a.get("/get-user/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(chi.URLParam(r, "id"))
		user, err := a.models.Users.Get(id)
		if err != nil {
			a.App.ErrorLog.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "User ID: %d | Name: %s %s | Email: %s\n", user.ID, user.FirstName, user.LastName, user.Email)
	})

	a.get("/update-user/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(chi.URLParam(r, "id"))
		user, err := a.models.Users.Get(id)
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}

		user.LastName = a.App.RandomString(10)
		validator := a.App.Validator(nil)
		user.LastName = ""
		user.Validate(validator)
		if !validator.Valid() {
			fmt.Fprintf(w, "Validation failed: %v", validator.Errors)
			return
		}

		err = user.Update(*user)
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "User ID: %d | Name: %s %s | Email: %s\n", user.ID, user.FirstName, user.LastName, user.Email)
	})

	fileServer := http.FileServer(http.FS(public.Public))
	a.App.Routes.Handle("/public/*", http.StripPrefix("/public/", fileServer))

	return a.App.Routes
}
