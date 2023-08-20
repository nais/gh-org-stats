package main

import (
	"fmt"

	"github.com/urfave/cli"
)

func HelloCommand() cli.Command {
	return cli.Command{
		Name:    "hello",
		Aliases: []string{"h"},
		Usage:   "say hello",
		Action: func(c *cli.Context) error {
			fmt.Println("Hello!")
			return nil
		},
	}
}

func GoodbyeCommand() cli.Command {
	return cli.Command{
		Name:    "goodbye",
		Aliases: []string{"g"},
		Usage:   "say goodbye",
		Action: func(c *cli.Context) error {
			fmt.Println("Goodbye!")
			return nil
		},
	}
}
