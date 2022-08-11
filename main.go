package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/guitmz/n26"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	fullDaysToLookup := 0

	cliArgs := os.Args[1:]
	if len(cliArgs) == 1 {
		days, err := strconv.ParseInt(cliArgs[0], 10, 32)
		if err != nil {
			panic("duhhh")
		}
		fullDaysToLookup = int(days)
	}

	endTime := n26.TimeStamp{Time: time.Now()}
	startTime := n26.TimeStamp{Time: endTime.Time.Add((-time.Hour * 24) * time.Duration(fullDaysToLookup))}

	fmt.Println("waiting for 2FA confirmation in app")
	client, err := n26.NewClient(n26.Auth{
		UserName:    os.Getenv("USERNAME"),
		Password:    os.Getenv("PASSWORD"),
		DeviceToken: os.Getenv("DEVICE_TOKEN"),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("auth complete")

	fmt.Println("auth woked")
	transactions, err := client.GetTransactions(startTime, endTime, "12")
	if err != nil {
		panic(err)
	}
	fmt.Println(transactions)
}
