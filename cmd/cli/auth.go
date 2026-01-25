package main

import (
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
)

func doAuth() error {

	// migrations
	dbType := cel.DB.DataType
	fileName := fmt.Sprintf("%d_%s.sql", time.Now().Unix(), "create_auth_tables")
	upFile := cel.RootPath + "/migrations/" + fileName + "." + dbType + ".up.sql"
	downFile := cel.RootPath + "/migrations/" + fileName + "." + dbType + ".down.sql"

	log.Println(dbType, upFile, downFile)

	err := copyFilefromTemplate("migrations/auth_tables."+dbType+".up.sql", upFile)
	if err != nil {
		exitGracefully(err)
	}

	err = copyDataToFile([]byte("drop table if exists users cascade; drop table if exists tokens cascade; drop table if exists remember_tokens cascade"), downFile)
	if err != nil {
		exitGracefully(err)
	}

	// run migrations
	err = doMigrate("up", "")
	if err != nil {
		exitGracefully(err)
	}

	// copy data models
	err = copyFilefromTemplate("data/user.go.txt", cel.RootPath+"/data/user.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFilefromTemplate("data/token.go.txt", cel.RootPath+"/data/token.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFilefromTemplate("data/remember_token.go.txt", cel.RootPath+"/data/remember_token.go")
	if err != nil {
		exitGracefully(err)
	}

	// copy middleware
	err = copyFilefromTemplate("middleware/auth.go.txt", cel.RootPath+"/middleware/auth.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFilefromTemplate("middleware/auth-token.go.txt", cel.RootPath+"/middleware/auth-token.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFilefromTemplate("middleware/remember.go.txt", cel.RootPath+"/middleware/remember.go")
	if err != nil {
		exitGracefully(err)
	}

	// copy handlers
	err = copyFilefromTemplate("handlers/auth-handler.go.txt", cel.RootPath+"/handlers/auth-handler.go")
	if err != nil {
		exitGracefully(err)
	}

	// copy mail template
	err = copyFilefromTemplate("mailer/password-reset.html.tmpl", cel.RootPath+"/mail/password-reset.html.tmpl")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFilefromTemplate("mailer/password-reset.plain.tmpl", cel.RootPath+"/mail/password-reset.plain.tmpl")
	if err != nil {
		exitGracefully(err)
	}

	// copy view templates
	err = copyFilefromTemplate("views/login.jet", cel.RootPath+"/views/login.jet")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFilefromTemplate("views/forgot.jet", cel.RootPath+"/views/forgot.jet")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFilefromTemplate("views/reset-password.jet", cel.RootPath+"/views/reset-password.jet")
	if err != nil {
		exitGracefully(err)
	}

	// after
	color.Yellow("- users, tokens and remember_tokens tables created")
	color.Yellow("- user and token models created")
	color.Yellow("- auth middleware created")
	color.Yellow("")
	color.Yellow("don't forget to add user and token models in data/models.go, and to appropriate middleware to your routes")

	return nil
}
