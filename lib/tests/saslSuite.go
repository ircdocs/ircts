// Copyright (c) 2019 Daniel Oaks <daniel@danieloaks.net>
// released under the MIT license

package tests

import (
	"encoding/base64"
	"fmt"

	"github.com/goshuirc/irc-go/ircmsg"
)

var saslTests TestGroup

func init() {
	saslLoginPlainTest := Test{
		Name:         "Login-PLAIN",
		Description:  "Tests the SASL PLAIN authentication method.",
		RequiredCaps: []string{"sasl"},
		Handler: func(name string, rm *RunManager) bool {
			// ensure we have a valid account
			haveValidAccount := 0 < len(rm.Config.Accounts)
			if !haveValidAccount {
				rm.Results.Set(name, NotApplicable, "No SASL PLAIN account (username+password) defined in config", "")
				return true
			}
			acc := rm.Config.Accounts[0]

			// make connection
			c, err := rm.Pool.NewConnection(rm.Config.Server)
			if err != nil {
				rm.Results.Set(name, Failure, fmt.Sprintf("Could not setup new connection: %s", err.Error()), "")
				return false
			}
			defer rm.Pool.DestroyConnection(c)

			c.SendSimpleMessage("CAP", "REQ", "sasl")
			c.SendSimpleMessage("NICK", rm.NewNick())
			c.SendSimpleMessage("USER", "t", "0", "*", name)

			c.SendSimpleMessage("AUTHENTICATE", "PLAIN")
			var authBytes []byte
			authBytes = append(authBytes, acc.Username...)
			authBytes = append(authBytes, '\000')
			authBytes = append(authBytes, acc.Username...)
			authBytes = append(authBytes, '\000')
			authBytes = append(authBytes, acc.Password...)
			c.SendSimpleMessage("AUTHENTICATE", base64.StdEncoding.EncodeToString(authBytes))
			c.SendSimpleMessage("CAP END")

			var got900, got903 bool
			for {
				line, err := c.GetLine()
				if err != nil {
					rm.Results.Set(name, Failure, fmt.Sprintf("Could not get reply line: %s", err.Error()), c.Traffic())
					return true
				}

				msg, err := ircmsg.ParseLine(line)
				if err != nil {
					rm.Results.Set(name, Failure, fmt.Sprintf("Failed to parse reply line: %s", err.Error()), c.Traffic())
					return true
				}

				if msg.Command == "001" {
					if got900 && got903 {
						rm.Results.Set(name, Success, "Successfully completed SASL PLAIN (username+password) authentication", c.Traffic())
					} else {
						rm.Results.Set(name, Failure, "RPL_WELCOME (001) encountered before SASL success numerics", c.Traffic())
					}
					return true
				}
				if msg.Command == "900" {
					got900 = true
				}
				if msg.Command == "903" {
					if !got900 {
						rm.Results.Set(name, Failure, "Got RPL_SASLSUCCESS (903) before or without getting RPL_LOGGEDIN (900)", c.Traffic())
						return true
					}
					got903 = true
				}
				for _, cmd := range []string{"901", "902", "904", "905", "906", "907", "908"} {
					if msg.Command == cmd {
						rm.Results.Set(name, Failure, "Got unexpected SASL failure or informational numeric", c.Traffic())
						return true
					}
				}
			}
		},
	}

	saslTests = TestGroup{
		Name:        "SASL",
		Description: "Tests SASL authentication methods.",
		Tests: []Test{
			saslLoginPlainTest,
		},
	}
}
