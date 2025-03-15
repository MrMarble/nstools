package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/mrmarble/nstools"
	"github.com/urfave/cli/v3"
)

func main() {
	var keys *nstools.Keys
	err := (&cli.Command{
		Name: "nstools",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "keys",
				Aliases: []string{"k"},
				Value:   "./prod.keys",
				Validator: func(s string) error {
					if _, err := os.Stat(s); err != nil {
						return err
					}
					f, err := os.ReadFile(s)
					if err != nil {
						return err
					}
					keys = nstools.NewKeys(f)
					return keys.Validate()
				},
			},
		},
		ArgsUsage: "[file to open]",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() == 0 {
				return cli.Exit("no file provided", 1)
			}
			c.String("keys")
			file, err := nstools.OpenFile(c.Args().First(), keys)
			if err != nil {
				return cli.Exit(err, 1)
			}

			switch f := file.(type) {
			case *nstools.PFS0:
				fmt.Printf("PFS0\nMagic:\t\t\t%s\nNumber of files:\t%d\nFiles:\n", f.Magic, len(f.Files))
				for _, file := range f.Files {
					offset := file.Offset - int64(f.HeaderSize)
					fmt.Printf("\t\t\tpfs0:/%s\t%s\t%08X-%08X\n", file.Name, humanize.Bytes(uint64(file.Size)), offset, offset+file.Size)
				}
				f.Close()
			case *nstools.CNMT:

			default:
				return cli.Exit("unknown file format", 1)
			}

			return nil
		},
	}).Run(context.Background(), os.Args)

	if err != nil {
		log.Fatal(err)
	}
}
