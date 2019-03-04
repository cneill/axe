package main

import (
	"bufio"
	_ "flag"
	"fmt"
	"log"
	"strings"

	"os"
)

// TODO: MAKE THIS FASTER; CALLING NEWPARSER/SCAN EVERY TIME IS DUMB; DON'T
// NEED TO RE-ASSIGN ITEMORDER, MAKE CHANNELS & STRUCTS, ETC
func parseSTDINLines(command string) {
	s := bufio.NewScanner(os.Stdin)
	i := 0
	for s.Scan() {
		parser := NewParser(s.Text())
		ll, err := parser.ParseLine()
		if err != nil {
			log.Fatalf("PARSER ERROR: %v", err)
		} else if ll == nil {
			continue
		}
		switch command {
		case "ips":
			fmt.Println(ll.IP)
		case "useragents":
			fmt.Println(ll.UserAgent)
		case "referers":
			fmt.Println(ll.Referer)
		default:
			fmt.Printf("%d: %+v\n", i, ll)
		}
		i++
	}
}

// TODO: make parsing happen at each line
func readSTDINLines() []string {
	s := bufio.NewScanner(os.Stdin)
	lines := []string{}
	for s.Scan() {
		lines = append(lines, s.Text())
	}
	return lines
}

func parseLine(line string) *LogLine {
	line = strings.TrimSpace(line)
	parser := NewParser(line)
	ll, err := parser.ParseLine()
	if err != nil {
		log.Fatalf("PARSER ERROR: %v", err)
	}
	return ll
}

func parseLines(lines []string) []*LogLine {
	ll := []*LogLine{}
	for _, line := range lines {
		ll = append(ll, parseLine(line))
	}
	return ll
}

func printLines(lines []string) {
	for i, line := range lines {
		fmt.Printf("%d: %s\n", i, line)
	}
}

func startCommand(input, args []string) error {
	if len(args) < 2 {
		printLines(input)
		return nil
	}

	var ll = parseLines(input)

	for i, l := range ll {
		if l == nil {
			continue
		}
		rawCmd := args[1]
		switch rawCmd {
		case "ips":
			fmt.Println(l.IP)
		case "useragents":
			fmt.Println(l.UserAgent)
		case "referers":
			fmt.Println(l.Referer)
		default:
			fmt.Printf("%d: %+v\n", i, l)
		}
	}

	return nil
}

func main() {
	/*
		input := readSTDINLines()
		if input == nil {
			return
		}

		if err := startCommand(input, os.Args); err != nil {
			log.Fatalf("failed to parse command: %v", err)
		}
	*/
	cmd := ""
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}
	parseSTDINLines(cmd)
}
