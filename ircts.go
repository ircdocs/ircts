// Copyright (c) 2019 Daniel Oaks <daniel@danieloaks.net>
// Copyright (c) 2019 Shivaram Lingamneni
// released under the MIT license

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/goshuirc/irc-go/ircmsg"

	"github.com/ircdocs/ircts/lib/tests"

	"github.com/docopt/docopt-go"
	"github.com/ircdocs/ircts/lib"
	"github.com/ircdocs/ircts/lib/conn"
	"github.com/ircdocs/ircts/lib/utils"
)

func main() {
	usage := `ircts.
ircts is an IRC Test Suite, designed to test software's compatibility with
accepted standards and common behaviour. To run it, simply setup the config
file to point it towards the server, and go!

Usage:
	ircts run [--config <file>] [--verbose]
	ircts -h | --help
	ircts --version
Options:
	--config <file>     Configuration file [default: ircts.yaml].
	-v --verbose       	Show more detail while running tests.
	-h --help          	Show this screen.
	--version          	Show version.`

	arguments, _ := docopt.ParseArgs(usage, nil, lib.SemVer)

	if arguments["run"].(bool) {
		// load config
		config, err := utils.LoadConfig(arguments["--config"].(string))
		if err != nil {
			log.Fatalln("Failed to load config:", err.Error())
		}

		// create connection list
		connectionPool := conn.NewConnectionPool()
		rm := tests.NewRunManager()
		rm.Pool = connectionPool
		rm.Config = config

		// mark all tests as N/A (not yet run)
		for _, TG := range tests.AllTests {
			for _, Test := range TG.Tests {
				name := fmt.Sprintf("%s-%s", TG.Name, Test.Name)

				rm.Results.Set(name, tests.NotApplicable, "Test cancelled and not run", "")
			}
		}

		// get supported caps
		c, err := rm.Pool.NewConnection(config.Server)
		if err != nil {
			log.Fatalln("Failed to connect to server for initial caps:", err.Error())
		}

		supportedCaps := make(map[string]bool)
		capValues := make(map[string]string)

		c.SendLine("CAP LS 302")
		for err == nil {
			line, err := c.GetLine()
			if err != nil {
				break
			}

			msg, err := ircmsg.ParseLine(line)
			if err != nil {
				break
			}

			if msg.Command == "CAP" {
				if 2 < len(msg.Params) && strings.ToUpper(msg.Params[1]) == "LS" {
					if msg.Params[2] == "*" {
						for _, fullVal := range strings.Fields(msg.Params[3]) {
							splitVal := strings.SplitN(fullVal, "=", 2)
							name := splitVal[0]
							var value string
							if 1 < len(splitVal) {
								value = splitVal[1]
							}
							supportedCaps[name] = true
							capValues[name] = value
						}
					} else {
						for _, fullVal := range strings.Fields(msg.Params[2]) {
							splitVal := strings.SplitN(fullVal, "=", 2)
							name := splitVal[0]
							var value string
							if 1 < len(splitVal) {
								value = splitVal[1]
							}
							supportedCaps[name] = true
							capValues[name] = value
						}
						c.SendLine("QUIT")
					}
				}
			}
		}

		// eventually we'll just return it nicely instead
		rm.Pool.DestroyConnection(c)

		// start running tests!
		for _, TG := range tests.AllTests {
			fmt.Println("")
			fmt.Println("Testing", TG.Name)
			for _, Test := range TG.Tests {
				fmt.Println("-", Test.Name)

				name := fmt.Sprintf("%s-%s", TG.Name, Test.Name)
				Test.Handler(name, rm)
			}
		}

		// print basic results
		fmt.Println("\n= Results =")
		for _, TG := range tests.AllTests {
			for _, Test := range TG.Tests {
				name := fmt.Sprintf("%s-%s", TG.Name, Test.Name)
				rm.Results.Print(name)
			}
		}
	}
}
