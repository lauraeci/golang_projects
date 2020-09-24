# Laura Unterguggenberger coding challenge May 2020 
loggregator - a go library which aggregates event logs from a server with logs of given format below 

# Details
Loggregator takes advantage of SQL functionality using a lightweight relational database sqlite and queries generated through the go ORM library https://gorm.io/ 

# To build
1. Copy project folder to $GOPATH.src/github.com
2. Run 'dep ensure'
3. Then 'go build' 
4. run with ./loggregator

Thank you,
-Laura

# Scenario

You are given a very simple log of events(`serverlog.csv`) for a server that allows users to upload or download files.


Using this log file as input, write a Go program that will output the following metrics:
1. How many users accessed the server?
2. How many uploads were larger than `50kB`?
3. How many times did `jeff22` upload to the server on April 15th, 2020?


# Details
The log is represented by a CSV formatted file where column data is ordered: `timestamp`, `username`, `operation`, `size`.

- The `timestamp` is recorded in the UNIX date format.
- The `username` is a unique identifier for the user.
- The `operation` indicates if an `upload` or `download` occurred, no other values will appear in this column.
- The `size` is an integer reflecting file size in `kB`.

You may use whatever packages you feel will be useful to solve this challenge.

# Important Note
The intention of this coding challenge is to gauge a candidate's ability to program in Go, write clean code, and discuss their work with others. While we would like to see your best work, we don't want candidates to spend a large amount of their free time optimizing their solution to this challenge. So, our expectation is not necessarily to receive an optimal or even completely correct solution. With this in mind, we recommend spending about an hour on this challenge.# ctrl_cmd_challenge
