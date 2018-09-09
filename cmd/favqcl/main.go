package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/marcsantiago/favqs"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "filter, f",
			Usage: "filter limits the quote catergories",
			Value: "science",
		},
		cli.Int64Flag{
			Name:  "limit, l",
			Usage: "the max number of quotes to return",
			Value: 1,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "single",
			Aliases: []string{"s"},
			Usage:   "prints the a random quote",
			Action:  printsQuoteOfDay,
		},
		{
			Name:    "many",
			Aliases: []string{"m"},
			Usage:   "prints a list of quotes filtered by -f and limited by -l",
			Action:  printsQuoteWithFilter,
		},
	}
	sort.Sort(cli.CommandsByName(app.Commands))
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func printsQuoteOfDay(c *cli.Context) error {
	client, err := favqs.New()
	if err != nil {
		return err
	}
	q, err := client.GetQuoteOfTheDay()
	if err != nil {
		return err
	}
	fmt.Printf("Author: %s\nQuote: %s\n", q.Quote.Author, q.Quote.Body)
	return nil
}

func printsQuoteWithFilter(c *cli.Context) error {
	client, err := favqs.New()
	if err != nil {
		return err
	}
	qs, err := client.GetQuotes(c.GlobalString("filter"), c.GlobalInt("limit"))
	if err != nil {
		return err
	}

	for _, q := range qs {
		fmt.Printf("Author: %s\nQuote: %s\n\n", q.Author, q.Body)
	}

	return nil
}
