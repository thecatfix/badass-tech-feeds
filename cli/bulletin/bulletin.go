package main

import (
	"bulletin/command"
	"bulletin/log"
	"bulletin/storage"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
)

const bulletinDir = ".bulletin"
const cacheBaseName = "cache"

func main() {
	err := mainErr()
	if err != nil {
		fmt.Printf("ERROR. %s\n", err)
		os.Exit(1)
	}
}

func mainErr() error {
	availableCommands := []string{
		command.ComposeCommandName,
		command.CleanCommandName,
		command.FetchCommandName,
		command.TestCommandName,
	}

	flag.Usage = func() {
		fmt.Printf("Available commands: %s\n", strings.Join(availableCommands, ", "))
		fmt.Printf("If commands are missing, it will run a default sequence generating a default bulletin.")
		flag.PrintDefaults()
	}
	var opts options
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	defaultCacheDir := path.Join(homeDir, bulletinDir, cacheBaseName)
	flag.StringVar(&opts.cacheDir, "cache", defaultCacheDir, "cache directory")
	flag.BoolVar(&opts.verbose, "v", false, "verbose log")
	flag.Parse()
	log.SetVerbose(opts.verbose)
	if opts.cacheDir == defaultCacheDir {
		if err := os.MkdirAll(defaultCacheDir, 0755); err != nil {
			return err
		}
	}

	storageInstance := &storage.Storage{
		Path: opts.cacheDir,
	}
	commands := make(map[string]command.Command)
	commands[command.FetchCommandName] = &command.FetchCommand{
		Storage: storageInstance,
	}
	commands[command.ComposeCommandName] = &command.ComposeCommand{
		Storage: storageInstance,
	}
	commands[command.TestCommandName] = &command.TestCommand{}
	commands[command.CleanCommandName] = &command.CleanCommand{
		Storage: storageInstance,
	}

	if flag.NArg() == 0 {
		cmd := command.RunSequence{
			Commands: []command.Command{
				commands[command.CleanCommandName],
				commands[command.FetchCommandName],
				commands[command.ComposeCommandName],
			}}
		return cmd.Execute([]string{})
	}
	if flag.NArg() >= 1 {
		commandString := flag.Arg(0)
		cmd, ok := commands[commandString]
		if !ok {
			return fmt.Errorf("unknown command: %s", commandString)
		}
		return cmd.Execute(flag.Args()[1:])
	}
	flag.Usage()
	return fmt.Errorf("pass only one command")

}

type options struct {
	cacheDir string
	verbose  bool
}
