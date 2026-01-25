package main

import (
	"errors"
	"os"
	"runtime/debug"
	"strings"

	"github.com/ansufw/celeritas"
	"github.com/fatih/color"
)

func getVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" {
			return info.Main.Version
		}
	}
	return "(devel)"
}

var version = getVersion()

var cel celeritas.Celeritas

func main() {

	var message string

	arg1, arg2, arg3, err := validateInput()
	if err != nil {
		exitGracefully(err)
	}

	setup(arg1, arg2)

	switch arg1 {
	case "help":
		showHelp()

	case "new":
		if arg2 == "" {
			exitGracefully(errors.New("new requires a subcommand"))
		}
		err = doNew(arg2)
		if err != nil {
			exitGracefully(err)
		}
		message = "New project created successfully"

	case "version":
		color.Yellow("Version: %s\n", version)

	case "migrate":
		if arg2 == "" {
			arg2 = "up"
		}
		err = doMigrate(arg2, arg3)
		if err != nil {
			exitGracefully(err)
		}
		message = "Migration completed successfully"

	case "make":
		if arg2 == "" {
			exitGracefully(errors.New("make requires a subcommand"))
		}
		err = doMake(arg2, arg3)
		if err != nil {
			exitGracefully(err)
		}
	default:
		showHelp()
		exitGracefully(errors.New("command is required"))
	}

	exitGracefully(nil, message)
}

func validateInput() (string, string, string, error) {
	var arg1, arg2, arg3 string

	if len(os.Args) > 1 {
		arg1 = strings.ToLower(os.Args[1])

		if len(os.Args) >= 3 {
			arg2 = strings.ToLower(os.Args[2])
		}

		if len(os.Args) >= 4 {
			arg3 = strings.ToLower(os.Args[3])
		}
	} else {
		color.Red("Error: command is required")
		showHelp()
		return "", "", "", errors.New("command is required")
	}

	return arg1, arg2, arg3, nil
}

func exitGracefully(err error, msg ...string) {
	message := ""
	if len(msg) > 0 {
		message = msg[0]
	}

	if err != nil {
		color.Red("Error: %v\n", err)
	}

	if len(message) > 0 {
		color.Yellow(message)
	} else {
		color.Green("Done!")
	}

	os.Exit(0)
}
