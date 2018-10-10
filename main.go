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

const maxRequest = 10

var (
	maxNumber int
	apiKey    string
	verbose   bool
	holders   *Holders
	client    *Client
	wg        *sync.WaitGroup
)

func init() {
	apiKey = "AQDH6DUI9THYHCS21XM81CEXS21DU14TI2"
	holders = NewHolders()
	wg = &sync.WaitGroup{}
}

func main() {
	// Set up flags.
	flag.StringVar(&apiKey, "k", "", "etherscan.io api key to perform all requests (leave it empty to use developer key)")
	flag.IntVar(&maxNumber, "n", 1000, "how many users to fetch")
	flag.BoolVar(&verbose, "v", false, "verbose output")
	flag.Usage = printHelp
	flag.Parse()
	// Check arguments.
	args := flag.Args()
	if len(args) < 2 {
		fmt.Println("Not enough arguments!")
		printHelp()
		os.Exit(0)
	}
	// Set token and file from args.
	address := args[0]

	// Init etherscan API client.
	client = NewClient(apiKey)

	// Start fetching process.
	log.Println("Start fetching transactions.")
	start := time.Now()
	page := 1
	offset := 3000
	done := false
	for !done {
		for i := 0; i < maxRequest; i++ {
			wg.Add(1)
			go fetchTransactions(address, page, offset, &done)
			page++
		}
		time.Sleep(time.Second * 5)
	}
	wg.Wait()

	// Print all holders to the file in csv format.
	if len(holders.m) > 0 {
		file := createFile(args[1])
		defer file.Close()
		for _, holder := range holders.m {
			fmt.Fprintf(file, "%s;%s;%d;%d\n", holder.Address, holder.Balance, holder.Transcation, holder.LastActive.Unix())
		}
	}
	end := time.Since(start)
	log.Printf("%d records stored.\n", len(holders.m))
	log.Printf("Done in %s.\n", end.String())
}

func fetchTransactions(address string, page, offset int, done *bool) {
	defer wg.Done()
	if verbose {
		log.Printf("Fetching %d page...\n", page)
	}
	txns, err := client.TokenTransferEvents(address, false, page, offset)
	if err != nil {
		log.Printf("%s\n", err)
		if err.Error() == "NOTOK" {
			*done = true
		} else {
			wg.Add(1)
			go fetchTransactions(address, page, offset, done)
		}
		return
	}
	// Page overflow its maximum.
	if len(txns) == 0 {
		*done = true
	} else {
		n := storeHolders(txns)
		log.Printf("Stored %3d / Common %d\n", n, len(holders.m))
		if len(holders.m) >= maxNumber {
			*done = true
		}
	}
}

func storeHolders(txns []Transcation) (n int) {
	// Set
	for _, txn := range txns {
		if (len(holders.m) + n) >= maxNumber {
			return
		}
		if _, ok := holders.Get(txn.From); !ok {
			timestamp, _ := strconv.ParseInt(txn.TimeStamp, 10, 64)
			holders.Set(txn.From, &Holder{
				Address:     txn.From,
				Transcation: 1,
				Balance:     "0",
				LastActive:  time.Unix(timestamp, 0),
			})
			n++
		}
	}

	return
}

func createFile(filename string) *os.File {
	// Set up file.
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0660)
	if err != nil {
		log.Fatalln(err)
	}
	// Truncate file.
	file.Truncate(0)
	file.Seek(0, 0)

	return file
}

func printHelp() {
	fmt.Printf("etherscan is a simple scanner tool for fetching cryptoholders data and exporting it into csv format.\n\n")
	fmt.Println("usage: etherscan [TOKEN] [FILE]")
	flag.PrintDefaults()
}
