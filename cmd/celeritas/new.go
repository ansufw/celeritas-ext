package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
)

var appURL string

func doNew(appName string) error {

	appName = strings.ToLower(appName)
	appURL = appName

	// sanitize the app name (convert url to single word)
	if strings.Contains(appName, "/") {
		exploded := strings.SplitAfter(appName, "/")
		appName = exploded[len(exploded)-1]
	}

	log.Println("Creating new app: " + appName)

	// git clone the skeleton app
	repoURL := "https://github.com/ansufw/celeritas-boilerplate.git"
	color.Green("\tCloning repository ", repoURL)
	_, err := git.PlainClone(appName, false, &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
		Depth:    1,
	})
	if err != nil {
		return err
	}

	// remove .git directory
	err = os.RemoveAll(fmt.Sprintf("./%s/.git", appName))
	if err != nil {
		return err
	}

	// create a ready to go .env file
	color.Yellow("\tCreating .env file...")
	data, err := templateFS.ReadFile("env.txt")
	if err != nil {
		return err
	}

	env := string(data)
	env = strings.ReplaceAll(env, "${APP_NAME}", appName)
	env = strings.ReplaceAll(env, "${KEY}", cel.RandomString(32))

	err = copyDataToFile([]byte(env), fmt.Sprintf("./%s/.env", appName))
	if err != nil {
		return err
	}

	// create a makefile
	if runtime.GOOS == "windows" {
		source, err := os.Open(fmt.Sprintf("./%s/Makefile.windows", appName))
		if err != nil {
			return err
		}
		defer source.Close()

		dest, err := os.Create(fmt.Sprintf("./%s/Makefile", appName))
		if err != nil {
			return err
		}
		defer dest.Close()

		_, err = io.Copy(dest, source)
		if err != nil {
			return err
		}
	} else {
		source, err := os.Open(fmt.Sprintf("./%s/Makefile.mac", appName))
		if err != nil {
			return err
		}
		defer source.Close()

		dest, err := os.Create(fmt.Sprintf("./%s/Makefile", appName))
		if err != nil {
			return err
		}
		defer dest.Close()

		_, err = io.Copy(dest, source)
		if err != nil {
			return err
		}
	}
	_ = os.Remove(fmt.Sprintf("./%s/Makefile.windows", appName))
	_ = os.Remove(fmt.Sprintf("./%s/Makefile.mac", appName))

	// update the go.mod file
	color.Yellow("\tCreating go.mod file...")
	_ = os.Remove("./" + appName + "/go.mod")

	data, err = templateFS.ReadFile("go.mod.txt")
	if err != nil {
		return err
	}

	goMod := string(data)
	goMod = strings.ReplaceAll(goMod, "${APP_NAME}", appName)

	err = copyDataToFile([]byte(goMod), fmt.Sprintf("./%s/go.mod", appName))
	if err != nil {
		return err
	}

	// update existing .go files with correct name/imports
	color.Yellow("\tUpdating source files...")
	os.Chdir("./" + appName)
	updateSource()

	// run go mod tidy in the project directory
	color.Yellow("\tRunning go mod tidy..")
	cmd := exec.Command("go", "mod", "tidy")
	err = cmd.Start()
	if err != nil {
		return err
	}

	color.Green("done building " + appURL)
	color.Green("go build something awesome!")

	return nil
}
