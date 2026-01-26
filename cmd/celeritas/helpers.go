package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

func showHelp() {
	color.Yellow(`Available commands:

	help                   - show this help message
	version                - print application version
	make migration <name>  - create new migration file
	make auth              - create and run migrations for authentication tables, and create models and middleware
	make handler <name>    - create a stub handler in the handlers directory
	make model <name>      - create a stub model in the models directory
	make mail <name>       - create two staters mail template in the mail directory
	make session           - create a table in the database as a session store
	migrate                - runs all up migrations that have not been run previously
	migrate down           - reverses the last migration
	migrate reset          - run down migrations and then run up migrations
	`)
}

func setup(arg1, arg2 string) {

	if arg1 != "new" && arg1 != "version" && arg1 != "help" {
		err := godotenv.Load()
		if err != nil {
			exitGracefully(err)
		}

		path, err := os.Getwd()
		if err != nil {
			exitGracefully(err)
		}

		cel.RootPath = path
		cel.DB.DataType = os.Getenv("DATABASE_TYPE")
	}

}

func getDSN() string {
	dbType := cel.DB.DataType
	if dbType == "pgx" {
		dbType = "postgres"
	}

	if dbType == "postgres" {
		var dsn string
		if os.Getenv("DATABASE_PASS") != "" {
			dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
				os.Getenv("DATABASE_USER"),
				os.Getenv("DATABASE_PASS"),
				os.Getenv("DATABASE_HOST"),
				os.Getenv("DATABASE_PORT"),
				os.Getenv("DATABASE_NAME"),
				os.Getenv("DATABASE_SSL_MODE"),
			)
		} else {
			dsn = fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=%s",
				os.Getenv("DATABASE_USER"),
				os.Getenv("DATABASE_HOST"),
				os.Getenv("DATABASE_PORT"),
				os.Getenv("DATABASE_NAME"),
				os.Getenv("DATABASE_SSL_MODE"),
			)
		}
		return dsn
	}

	return "mysql://" + cel.BuildDSN()
}

func updateSourceFiles(path string, fi os.FileInfo, err error) error {

	// check for an error before doing anything else
	if err != nil {
		return err
	}

	// check if current file is directory
	if fi.IsDir() {
		return nil
	}

	// only check go files
	matched, err := filepath.Match("*.go", fi.Name())
	if err != nil {
		return err
	}

	// we have matching file
	if matched {
		// read file content
		read, err := os.ReadFile(path)
		if err != nil {
			exitGracefully(err)
		}

		newContents := strings.Replace(string(read), "boilerplate", appURL, -1)

		// write new contents to file
		err = os.WriteFile(path, []byte(newContents), 0)
		if err != nil {
			exitGracefully(err)
		}
	}

	return nil
}

func updateSource() {
	// walk entire project folder, including subfolders
	err := filepath.Walk(".", updateSourceFiles)
	if err != nil {
		exitGracefully(err)
	}
}
