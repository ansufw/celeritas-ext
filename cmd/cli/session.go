package main

import (
	"fmt"
	"time"
)

func doSession() error {

	dbType := cel.DB.DataType
	if dbType == "mariadb" {
		dbType = "mysql"
	}

	if dbType == "postgresql" {
		dbType = "postgres"
	}

	fileName := fmt.Sprintf("%d_%s.sql", time.Now().Unix(), "create_session_table")

	upFile := cel.RootPath + "/migrations/" + fileName + "." + dbType + ".up.sql"
	downFile := cel.RootPath + "/migrations/" + fileName + "." + dbType + ".down.sql"

	err := copyFilefromTemplate("migrations/"+dbType+"_session.sql", upFile)
	if err != nil {
		exitGracefully(err)
	}

	err = copyDataToFile([]byte("drop table sessions"), downFile)
	if err != nil {
		exitGracefully(err)
	}

	err = doMigrate("up", "")
	if err != nil {
		exitGracefully(err)
	}

	return nil
}
