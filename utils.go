package main

import (
	"flag"
	"fmt"
	"math/big"
)

func Add(s1, s2 string) string {
	i1 := &big.Int{}
	i2 := &big.Int{}
	i1.SetString(s1, 10)
	i2.SetString(s2, 10)

	return i1.Add(i1, i2).Text(10)
}

func Sub(s1, s2 string) string {
	i1 := &big.Int{}
	i2 := &big.Int{}
	i1.SetString(s1, 10)
	i2.SetString(s2, 10)

	return i1.Sub(i1, i2).Text(10)
}

func PrintHelp() {
	fmt.Printf("etherscan is a simple scanner tool for fetching cryptoholders data and exporting it into csv format.\n\n")
	fmt.Println("usage: etherscan [TOKEN] [FILE]")
	flag.PrintDefaults()
}
