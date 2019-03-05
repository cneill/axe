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
var (
	ipsFlagset, uaFlagset, refFlagset *flag.FlagSet

	ipsResolve bool

	uaSimplify bool
)

func init() {
	errHandle := flag.ExitOnError

	ipsFlagset = flag.NewFlagSet("ips", errHandle)
	ipsFlagset.BoolVar(&ipsResolve, "resolve", false, "resolve IPs to hostnames where possible")

	uaFlagset = flag.NewFlagSet("user-agents", errHandle)
	uaFlagset.BoolVar(&uaSimplify, "simplify", false, "simplify user-agents")

	refFlagset = flag.NewFlagSet("referers", errHandle)

}

func defaultPrintFunc(l *LogLine) {
	fmt.Println(l.String())
}

func defaultErrFunc(err error) {
	_, printErr := fmt.Fprintf(os.Stderr, "%v\n", err)
	if printErr != nil {
		log.Fatalf("SUPER SERIOUS: %v: %v", printErr, err)
	}
}

func parseCLI(args []string) llFunc {
	cmd := ""
	if len(args) > 1 {
		cmd = args[1]
	}

	flag.Usage = func() {
		fmt.Println("axe takes logfiles as STDIN and prints the information requested")
		fmt.Printf("Usage: %s [command]\n", args[0])
		fmt.Println("Commands:")
		fmt.Println("\tips")
		fmt.Println("\tpaths")
		fmt.Println("\treferers")
		fmt.Println("\trequests")
		fmt.Println("\ttimes")
		fmt.Println("\tuseragents")

		flag.PrintDefaults()
	}

	flag.Parse()

	switch cmd {
	case "":
		return defaultPrintFunc
	case "help":
		flag.Usage()
		os.Exit(0)
	case "ips":
		return func(l *LogLine) {
			fmt.Println(l.IP.String())
		}
	case "paths":
		return func(l *LogLine) {
			if l.Request != nil {
				fmt.Println(l.Request.URL.String())
			}
		}
	case "referers":
		return func(l *LogLine) {
			fmt.Println(l.Referer.String())
		}
	case "requests":
		return func(l *LogLine) {
			if l.Request != nil {
				fmt.Printf("%s %s %s\n", l.Request.Method, l.Request.URL.String(), l.Request.Proto)
			}
		}
	case "times":
		return func(l *LogLine) {
			fmt.Println(l.Time)
		}
	case "useragents":
		return func(l *LogLine) {
			fmt.Println(l.UserAgent)
		}
	default:
		log.Fatalf("unknown option: %s", cmd)
	}
	return defaultPrintFunc
}

func main() {
	printFunc := parseCLI(os.Args)
	axe := NewAxe(1, printFunc, defaultErrFunc)
	axe.Start()
}
