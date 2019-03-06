package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

// TODO: allow parsing from file
// TODO: "fast" mode - don't care about interleaving lines, more workers
// TODO: ability to suppress errors?
func defaultPrintFunc(l *LogLine) {
	fmt.Println(l.String())
}

func defaultErrFunc(err error) {
	_, printErr := fmt.Fprintf(os.Stderr, "%v\n", err)
	if printErr != nil {
		log.Fatalf("SUPER SERIOUS: %v: %v", printErr, err)
	}
}

var cmdList = commands{}

func parseCLI(args []string) llFunc {
	cmd := ""
	cmdArgs := []string{}

	if len(args) == 1 {
		return defaultPrintFunc
	} else if len(args) > 1 {
		cmd = args[1]
		if len(args) > 2 {
			cmdArgs = args[2:]
		}
	}

	flag.Parse()

	if c := cmdList.find(cmd); c != nil {
		pf, err := c.execute(cmdArgs)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		return pf
	}

	flag.Usage()
	fmt.Printf("Command not found: %s\n", cmd)
	os.Exit(1)

	return nil
}

func main() {
	printFunc := parseCLI(os.Args)
	axe := NewAxe(1, printFunc, defaultErrFunc)
	axe.Start()
}
