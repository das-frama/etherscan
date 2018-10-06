package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

const offset = 1000

var holders *Holders

func main() {
	// Set up flags.
	apiKey := flag.String("k", "", "etherscan.io api key to perform all requests (leave it empty to use developer key)")
	flag.Usage = PrintHelp
	flag.Parse()
	// Check arguments.
	args := flag.Args()
	if len(args) < 2 {
		fmt.Println("Not enough arguments!")
		PrintHelp()
		os.Exit(0)
	}
	// Set token and file from args.
	contractAddress := args[0]
	file, err := os.OpenFile(args[1], os.O_CREATE|os.O_WRONLY, 0660)
	if err != nil {
		log.Fatalln(err)
		os.Exit(-1)
	}
	defer file.Close()
	// Truncate file.
	file.Truncate(0)
	file.Seek(0, 0)

	// Set ApiKey if non is provided.
	if *apiKey == "" {
		*apiKey = "AQDH6DUI9THYHCS21XM81CEXS21DU14TI2"
	}
	// Set etherscan client.
	c := NewClient(*apiKey)

	// Init variables.
	holders = NewHolders()
	wg := &sync.WaitGroup{}
	page := 1
	done := false
	start := time.Now()

	log.Println("Start fetching transactions. It should take just a couple of minutes or so.")
	for !done {
		wg.Add(1)
		go fetchTransactions(c, contractAddress, page, &done, wg)
		page++
		time.Sleep(time.Second * 1)
	}

	wg.Wait()
	// Print all holders to the file in csv format.
	for _, holder := range holders.m {
		fmt.Fprintf(file, "%s;%s;%d;%d\n", holder.Address, holder.Balance, holder.Transcation, holder.LastActive.Unix())
	}
	end := time.Since(start)
	log.Printf("Done in %s.\n", end.String())
}

func fetchTransactions(c *Client, contractAddress string, page int, done *bool, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Printf("Fetching %d page...\n", page)
	txns, err := c.TokenTransferEvents(contractAddress, "", true, page, offset)
	if err != nil {
		log.Fatalln(err)
		return
	}
	// Page overflow its maximum.
	if len(txns) == 0 {
		*done = true
	} else {
		storeTransactions(txns)
	}
}

func storeTransactions(txns []Transcation) {
	// Set
	for _, txn := range txns {
		timestamp, err := strconv.ParseInt(txn.TimeStamp, 10, 64)
		if err != nil {
			log.Println(err)
		}
		if holderFrom, ok := holders.Get(txn.From); ok {
			holderFrom.Transcation++
			holderFrom.Balance = Sub(holderFrom.Balance, txn.Value)
		} else {
			holders.Set(txn.From, &Holder{
				Address:     txn.From,
				Transcation: 1,
				Balance:     "-" + txn.Value,
				LastActive:  time.Unix(timestamp, 0),
			})
		}
		if holderTo, ok := holders.Get(txn.To); ok {
			holderTo.Transcation++
			holderTo.Balance = Add(holderTo.Balance, txn.Value)
		} else {
			holders.Set(txn.To, &Holder{
				Address:     txn.To,
				Transcation: 1,
				Balance:     txn.Value,
				LastActive:  time.Unix(timestamp, 0),
			})
		}
	}
}
