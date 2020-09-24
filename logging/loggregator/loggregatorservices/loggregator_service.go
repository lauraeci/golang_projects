package loggregatorservices

import (
	"log"
	"time"

	"github.com/jinzhu/gorm"
)

// LoggregatorService maintains operations on Log
type LoggregatorService struct {
	DB *gorm.DB
}

// Log representing a CSV formatted file from a server that allows users to upload or download files
type Log struct {
	gorm.Model
	// chose not to format the database timestamp in UNIX format because time.Time has better type checking, someone could put in a float64 into the db that's not a valid time
	Timestamp time.Time `json:"timestamp"`// gorm supports defaults
	Username  string    `json:"username"`
	Operation string    `json:"operation"`
	Size      int       `json:"size"`
}

// FindBy find logs by Log attributes and or size Limit and or day
func (s *LoggregatorService) FindBy(query Log, sizeLimit int, day string) []Log {
	var serverLogs []Log
	db := s.query(query, sizeLimit, day)
	db.Find(&serverLogs)
	return serverLogs
}

// UserCount get the number of unique users that accessed the server
// params: filter by day
func (s *LoggregatorService) UserCount(day string) (int, error) {
	var count int
	db := s.DB.Model(&Log{})
	if day != "" {
		db = db.Where("DATE(timestamp) = ?", day)
	}
	db = db.Model(&Log{}).Select("COUNT(DISTINCT username)").Count(&count)
	return count, nil
}

// CountBy count logs by Log attributes and or size Limit
func (s *LoggregatorService) CountBy(query Log, sizeLimit int, day string) (int, error) {
	var count int
	db := s.query(query, sizeLimit, day)
	db.Count(&count)
	return count, nil
}

// Create server log records
func (s *LoggregatorService) Create(serverLog []Log) {
	for _, l := range serverLog {
		s.DB.Create(&l)
	}
}

// UploadsGreaterThan50k get uploads greater than 50k
func (s *LoggregatorService) UploadsGreaterThan50k() {
	// Uploads over 50kb
	limit := 50
	query := Log{
		Operation: "upload",
	}
	uploadCount, err := s.CountBy(query, limit, "")
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("How many uploads were larger than `50kB`? %v", uploadCount)
	}
}

// UploadsForUserForDate get uploads for user by day
func (s *LoggregatorService) UploadsForUserForDate() {
	// How many times did a user access the server in one day
	query := Log{
		Operation: "upload",
		Username:  "jeff22",
	}
	uploadCount, err := s.CountBy(query, 0, "2020-04-15")
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("How many times did `jeff22` upload to the server on April 15th, 2020? %v", uploadCount)
	}
}

// UniqueUsers get unique users who accessed the server
// param: day filter by day
func (s *LoggregatorService) UniqueUsers(day string) {
	// Unique users accessing the server
	userCount, err := s.UserCount(day)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("How many users accessed the server? %+v", userCount)
	}
}

// private

func (s *LoggregatorService) query(query Log, sizeLimit int, day string) *gorm.DB {
	db := s.DB.Model(&Log{})
	if sizeLimit > 0 {
		db = db.Where("size > ?", sizeLimit)
	}
	if day != "" {
		db = db.Where("DATE(timestamp) = ?", day)
	}
	db = db.Where(query)
	return db
}
