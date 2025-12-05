package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:  "test",
				Usage: "use config to run tests",
				Action: func(context.Context, *cli.Command) error {
					return runTests()
				},
			},
			{
				Name:  "upload",
				Usage: "use config to send scripts to upload iflow scripts",
				Action: func(context.Context, *cli.Command) error {
					return runUploads()
				},
			},
		},
		Name:  "inco",
		Usage: "make groovy script manipulation easy",
		Action: func(context.Context, *cli.Command) error {
			fmt.Println("inco !")
			return nil
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
