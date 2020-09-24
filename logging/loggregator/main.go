package main

import (
	"encoding/csv"
	"fmt"
	loggregator "github.com/lauraeci/golang_projects/logging/loggregator/loggregatorservices"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jinzhu/gorm" // https://gorm.io/docs/
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// Formats for parsing time in the and day strings into time.Time
const (
	// Almost UnixDate but not quite "Mon Jan _2 15:04:05 MST 2006" day is 02 vs _2
	ServerTimeStamp = "Mon Jan 02 15:04:05 MST 2006"
	Day             = "Jan 02 2006"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Unix and Size could be interface methods

// Unix converts a string timestamp from the server
func Unix(timestamp string) time.Time {
	t, err := time.Parse(ServerTimeStamp, timestamp)
	check(err)
	return t
}

// Size converts a string size to an int size
func Size(size string) int {
	s, err := strconv.Atoi(size)
	check(err)
	return s
}

// simple implementation just moves the file after reading it so we don't import the data more than once
func logRotate(filename string) {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return
	}
	src := filename
	dest := fmt.Sprintf("%v%v", src, time.Now().Unix())
	err = os.Rename(src, dest)
	if err != nil {
		log.Print(err)
		return
	}
}

// ParseLogs a csv file filename into loggregator.Log
func ParseLogs(filename string) ([]loggregator.Log, error) {
	var serverLogs []loggregator.Log
	// could also check the presence of a previously read file to make sure the logs are still streaming correctly otherwise the operator has to manually look
	fileInfo, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("hopefully file %v was already read", filename)
	}
	log.Printf("creating logs from %v", fileInfo)

	csvFile, err := os.Open(filename)

	if err != nil {
		return nil, err
	}

	defer csvFile.Close()

	csvReader := csv.NewReader(csvFile)

	for {
		row, err := csvReader.Read()

		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		// Hack for "encoding/csv": don't try to create a log out of the headers of the first line.
		// It would have been nice if the library had a better treatment for the header and didn't consider it one of the rows or something
		if row[0] == "timestamp" {
			continue
		}

		serverLog := loggregator.Log{
			Timestamp: Unix(row[0]),
			Username:  row[1],
			Operation: row[2],
			Size:      Size(row[3]),
		}
		serverLogs = append(serverLogs, serverLog)
	}
	return serverLogs, err
}

func main() {
	// The down side to using structured data in a relational database is if the format of the logs changes then this would break. If the format is changing frequently, could switch to something like MongoDB.
	// Another good option is to stream to a cloud service like Loggly or if you're on AWS, CloudWatch but counting unique users might be difficult to do.
	// Gorm allows me to change databases easily without rewriting everything.
	// Sqlite3 doesn't require any set up unlike MySQL, MariaDB, Postgres, etc. Handles operations much more efficiently than I could come up with myself... pages and btrees? Julia Evans has a great post on this...
	// Sqlite3: https://jvns.ca/blog/2014/09/27/how-does-sqlite-work-part-1-pages/
	db, err := gorm.Open("sqlite3", "loggregator_dev.db")
	logMode := true

	// Set to true to see database logs
	db.LogMode(logMode)

	if err != nil {
		panic("failed to connect database")
	}

	defer db.Close()

	loggregatorService := &loggregator.LoggregatorService{
		DB: db,
	}

	// Migrate the schema
	// Adding a new column is as easy as adding a new field to the Log Struct but may require to drop the old table or field if there are name changes. Additions are handled well.
	db.AutoMigrate(&loggregator.Log{})

	// for the sake of this challenge, just parsed the csv file but there are probably ways to import the file as a csv directly into the db
	filename := "server_log.csv"
	serverLogs, err := ParseLogs(filename)
	// Having an error is okay if we already read the file and it was rotated, but could probably be handled better.
	if err != nil && logMode {
		msg := fmt.Sprintf("no logs found: %v", err)
		log.Printf(msg)
	}

	loggregatorService.Create(serverLogs)

	// Rotate the log so we don't upload the same logs multiple times.
	// Kept this simple to avoid repeated records. There are probably some different ways to do this to make sure no logs are lost depending on how the log files are streamed.
	logRotate(filename)

	// the functionality of these methods could also be provided in a public api over https would return queries that do what these hardcoded examples do instead of importing a library.
	// example: running a golang Microservice using: https://github.com/gorilla/mux, too heavy handed for this programming challenge
	loggregatorService.UniqueUsers("")
	loggregatorService.UploadsGreaterThan50k()
	loggregatorService.UploadsForUserForDate()
}
