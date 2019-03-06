package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
)

func init() {
	errHandle := flag.ExitOnError

	helpFS := flag.NewFlagSet("help", errHandle)
	newCommand(helpFS, nil, func(args []string) error {
		if len(args) == 0 {
			flag.Usage()
			os.Exit(0)
		}

		if c := cmdList.find(args[0]); c != nil {
			c.fs.Usage()
			os.Exit(0)
		}

		return fmt.Errorf("command not found: %s", args[0])
	})

	ipsFS := flag.NewFlagSet("ips", errHandle)
	ipsFS.Bool("resolve", false, "resolve IPs to hostnames where possible")
	newCommand(ipsFS, func(ll *LogLine) {
		fmt.Println(ll.IP.String())
	})

	pathsFS := flag.NewFlagSet("paths", errHandle)
	newCommand(pathsFS, func(ll *LogLine) {
		if ll.Request != nil && ll.Request.URL != nil {
			fmt.Println(ll.Request.URL.String())
		}
	})

	reqsFS := flag.NewFlagSet("requests", errHandle)
	newCommand(reqsFS, func(ll *LogLine) {
		if ll.Request != nil && ll.Request.URL != nil {
			fmt.Printf("%s %s %s\n", ll.Request.Method, ll.Request.URL.String(), ll.Request.Proto)
		}
	})

	refsFS := flag.NewFlagSet("referers", errHandle)
	newCommand(refsFS, func(ll *LogLine) {
		if ll.Referer != nil {
			fmt.Println(ll.Referer.String())
		}
	})

	statsFS := flag.NewFlagSet("statuses", errHandle)
	newCommand(statsFS, func(ll *LogLine) {
		fmt.Println(ll.Status)
	})

	timesFS := flag.NewFlagSet("times", errHandle)
	timesFS.String("format", nginxTimeFormat, "specify the format for time output")
	newCommand(timesFS, func(ll *LogLine) {
		fmt.Printf("[%s]\n", ll.Time.Format(nginxTimeFormat))
	})

	uaFS := flag.NewFlagSet("user-agents", errHandle)
	uaFS.Bool("simplify", false, "simplify user-agents")
	newCommand(uaFS, func(ll *LogLine) {
		fmt.Println(ll.UserAgent)
	})

	flag.Usage = func() {
		fmt.Println(cmdList.usageStr())
		flag.PrintDefaults()
	}
}

type execFunc func(args []string) error
type command struct {
	name string
	fs   *flag.FlagSet // flag set
	ef   execFunc      // exec function - done before returning print func
	pf   llFunc        // print func
}

func newCommand(fs *flag.FlagSet, pf llFunc, ef ...execFunc) *command {
	c := &command{
		name: fs.Name(),
		fs:   fs,
		pf:   pf,
	}

	if len(ef) > 0 {
		c.ef = ef[0]
	}

	cmdList = append(cmdList, c)
	return c
}

func (c *command) execute(args []string) (llFunc, error) {
	err := c.fs.Parse(args)
	if err != nil {
		return nil, err
	}

	if err := c.ef(args); err != nil {
		return nil, fmt.Errorf("%s: %v", c.name, err)
	}

	return c.pf, nil
}

type commands []*command

func (c commands) find(name string) *command {
	for _, cmd := range c {
		if cmd.name == name {
			return cmd
		}
	}
	return nil
}

func (c commands) usageStr() string {
	usage := "axe takes logfiles as STDIN and prints the information requested\n"
	usage += "Usage: axe [command] [options]\n"
	usage += "\nCommands and options:\n"

	for _, cmd := range c {
		usage += fmt.Sprintf("%s\n", cmd.name)

		// grab the output from PrintDefaults, which I don't want to rewrite
		buf := bytes.NewBufferString("")
		cmd.fs.SetOutput(buf)
		cmd.fs.PrintDefaults()
		defStr := buf.String()
		if defStr != "" {
			usage += buf.String()
		}
	}

	return usage
}
