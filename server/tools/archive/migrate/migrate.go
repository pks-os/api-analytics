package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tom-draper/api-analytics/server/database"

	"github.com/joho/godotenv"
	supa "github.com/nedpals/supabase-go"
	"github.com/oschwald/geoip2-golang"
)

func getSupabaseLogin() (string, string) {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	return supabaseURL, supabaseKey
}

func getCountryCode(db *geoip2.Reader, IPAddress string) (string, error) {
	var location string
	if IPAddress != "" {
		ip := net.ParseIP(IPAddress)
		record, err := db.Country(ip)
		if err != nil {
			return location, err
		}
		location = record.Country.IsoCode
	}
	return location, nil
}

func FixUserAgents() {
	f, err := os.Open("supabase_tyirpladmhanzkwhmspj_New Query.csv")
	if err != nil {
		log.Fatal("Unable to read input file ", err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	csvReader.Comma = '~'
	csvReader.LazyQuotes = true
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for ", err)
	}

	// "user_agent","method","created_at","path","api_key","status","response_time","request_id","framework","hostname","ip_address","location"
	var query strings.Builder
	query.WriteString("UPDATE requests as r SET user_agent = r2.user_agent from (values ")
	for i, record := range records {
		var userAgent string = record[0]
		if len(userAgent) > 255 {
			userAgent = userAgent[:255]
		}
		query.WriteString(fmt.Sprintf("(%d, '%s')", i, userAgent))
		if i < len(records)-1 {
			query.WriteString(",")
		}
	}

	query.WriteString(") as r2(request_id, user_agent) where r2.request_id = r.request_id;")

	// fmt.Println(query.String())
	db := database.OpenDBConnection()
	_, err = db.Query(query.String())
	if err != nil {
		panic(err)
	}
	db.Close()
}

func readRequests() []database.RequestRow {
	f, err := os.Open("supabase_tyirpladmhanzkwhmspj_New Query.csv")
	if err != nil {
		log.Fatal("Unable to read input file ", err)
	}
	defer f.Close()

	db, err := geoip2.Open("GeoLite2-Country.mmdb")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	csvReader := csv.NewReader(f)
	csvReader.Comma = '~'
	csvReader.LazyQuotes = true
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for ", err)
	}

	// "user_agent","method","created_at","path","api_key","status","response_time","request_id","framework","hostname","ip_address","location"
	result := make([]database.RequestRow, 0)
	for _, record := range records {
		r := new(database.RequestRow)
		r.UserAgent = record[0]
		method, _ := strconv.Atoi(record[1])
		r.Method = int16(method)
		for _, format := range []string{"2006-01-02 15:04:05.000000+00", "2006-01-02 15:04:05.00000+00", "2006-01-02 15:04:05.0000+00", "2006-01-02 15:04:05.000+00", "2006-01-02 15:04:05.00+00", "2006-01-02 15:04:05.0+00"} {
			r.CreatedAt, err = time.Parse(format, record[2])
			if err == nil {
				break
			}
		}
		r.Path = record[3]
		r.APIKey = record[4]
		status, _ := strconv.Atoi(record[5])
		r.Status = int16(status)
		responseTime, _ := strconv.Atoi(record[6])
		r.ResponseTime = int16(responseTime)
		r.RequestID, _ = strconv.Atoi(record[7])
		framework, _ := strconv.Atoi(record[8])
		r.Framework = int16(framework)
		r.Hostname = record[9]
		r.IPAddress = record[10]
		loc, _ := getCountryCode(db, record[10])
		r.Location = loc
		result = append(result, *r)
	}

	return result
}

func MigrateSupabaseRequests() {
	result := readRequests()

	var db *sql.DB
	var query strings.Builder
	i := 0
	for _, request := range result {
		if i > 0 {
			query.WriteString(",")
		} else {
			query = strings.Builder{}
			query.WriteString("INSERT INTO requests (api_key, method, created_at, path, status, response_time, framework, hostname, ip_address, location) VALUES")
		}
		if request.IPAddress == "" || request.IPAddress == "testclient" {
			query.WriteString(fmt.Sprintf(" ('%s', %d, '%s', '%s', %d, %d, %d, '%s', NULL, '%s')", request.APIKey, request.Method, request.CreatedAt.UTC().Format(time.RFC3339), request.Path, request.Status, request.ResponseTime, request.Framework, request.Hostname, request.Location))
		} else {
			query.WriteString(fmt.Sprintf(" ('%s', %d, '%s', '%s', %d, %d, %d, '%s', '%s', '%s')", request.APIKey, request.Method, request.CreatedAt.UTC().Format(time.RFC3339), request.Path, request.Status, request.ResponseTime, request.Framework, request.Hostname, request.IPAddress, request.Location))
		}

		i++
		if i == 10000 {
			query.WriteString(";")

			fmt.Println("Write to database")

			db = database.OpenDBConnection()
			_, err := db.Query(query.String())
			if err != nil {
				panic(err)
			}
			i = 0
			db.Close()
			time.Sleep(10 * time.Second)
		}
	}

	db = database.OpenDBConnection()
	query.WriteString(";")

	fmt.Println("Final write to database")

	_, err := db.Query(query.String())
	if err != nil {
		panic(err)
	}
	fmt.Println("Complete")
}

func MigrateSupabaseUsers(db *sql.DB, supabase *supa.Client) {
	var result []database.UserRow
	err := supabase.DB.From("Users").Select("*").Execute(&result)
	if err != nil {
		panic(err)
	}

	var query strings.Builder
	query.WriteString("INSERT INTO users (api_key, user_id, created_at) VALUES")
	for i, user := range result {
		if i > 0 {
			query.WriteString(",")
		}
		query.WriteString(fmt.Sprintf(" ('%s', '%s', '%s')", user.APIKey, user.UserID, user.CreatedAt.UTC().Format(time.RFC3339)))
	}
	query.WriteString(";")

	_, err = db.Query(query.String())
	if err != nil {
		panic(err)
	}
}

func MigrateSupabaseMonitors(db *sql.DB, supabase *supa.Client) {
	var result []database.MonitorRow
	err := supabase.DB.From("Monitor").Select("*").Execute(&result)
	if err != nil {
		panic(err)
	}

	var query strings.Builder
	query.WriteString("INSERT INTO monitor (api_key, url, secure, ping, created_at) VALUES")
	for i, monitor := range result {
		if i > 0 {
			query.WriteString(",")
		}
		query.WriteString(fmt.Sprintf(" ('%s', '%s', %t, %t, '%s')", monitor.APIKey, monitor.URL, monitor.Secure, monitor.Ping, monitor.CreatedAt.UTC().Format(time.RFC3339)))
	}
	query.WriteString(";")

	_, err = db.Query(query.String())
	if err != nil {
		panic(err)
	}
}

func MigrateSupabasePings(db *sql.DB, supabase *supa.Client) {
	var result []database.PingsRow
	err := supabase.DB.From("Pings").Select("*").Execute(&result)
	if err != nil {
		panic(err)
	}

	var query strings.Builder
	query.WriteString("INSERT INTO pings (api_key, url, response_time, status, created_at) VALUES")
	for i, monitor := range result {
		if i > 0 {
			query.WriteString(",")
		}
		query.WriteString(fmt.Sprintf(" ('%s', '%s', %d, %d, '%s')", monitor.APIKey, monitor.URL, monitor.ResponseTime, monitor.Status, monitor.CreatedAt.UTC().Format(time.RFC3339)))
	}
	query.WriteString(";")

	_, err = db.Query(query.String())
	if err != nil {
		panic(err)
	}
}

func MigrateSupabaseData() {
	// supabaseURL, supabaseKey := getSupabaseLogin()
	// supabase := supa.CreateClient(supabaseURL, supabaseKey)

	// db := OpenDBConnection()

	MigrateSupabaseRequests()
}
