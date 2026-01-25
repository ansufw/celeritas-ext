package main

import (
	"errors"
	"os"

	"github.com/ansufw/celeritas/cmd/celeritas/templates"
)

var templateFS = templates.Templates

func copyFilefromTemplate(templatePath string, destinationPath string) error {

	if fileExists(destinationPath) {
		return errors.New(destinationPath + " file already exists")
	}

	data, err := templateFS.ReadFile(templatePath)
	if err != nil {
		exitGracefully(err)
	}

	err = copyDataToFile(data, destinationPath)
	if err != nil {
		exitGracefully(err)
	}

	return nil
}

func copyDataToFile(data []byte, to string) error {
	return os.WriteFile(to, data, 0644)
}

func fileExists(fileToCheck string) bool {
	if _, err := os.Stat(fileToCheck); os.IsNotExist(err) {
		return false
	}
	return true
}
