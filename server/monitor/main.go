package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tom-draper/api-analytics/server/database"
)

func getMethod(ping bool) string {
	var method string
	if ping {
		method = "HEAD"
	} else {
		method = "GET"
	}
	return method
}

func ping(client http.Client, url string, secure bool, ping bool) (int, time.Duration, error) {
	method := getMethod(ping)

	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return 0, time.Duration(0), err
	}

	// Make request
	start := time.Now()
	response, err := client.Do(request)
	if err != nil {
		return 0, time.Duration(0), err
	}
	elapsed := time.Since(start)

	response.Body.Close()

	return response.StatusCode, elapsed, nil
}

func getMonitoredURLs(conn *pgx.Conn) []database.MonitorRow {
	query := "SELECT * FROM monitor;"
	rows, err := conn.Query(context.Background(), query)
	if err != nil {
		panic(err)
	}

	// Read monitors into list to return
	monitors := make([]database.MonitorRow, 0)
	for rows.Next() {
		monitor := new(database.MonitorRow)
		err := rows.Scan(&monitor.APIKey, &monitor.URL, &monitor.Secure, &monitor.Ping, &monitor.CreatedAt)
		if err == nil {
			monitors = append(monitors, *monitor)
		}
	}

	return monitors
}

func deleteExpiredPings(conn *pgx.Conn) {
	query := fmt.Sprintf("DELETE FROM pings WHERE created_at < '%s';", time.Now().Add(-60*24*time.Hour).UTC().Format("2006-01-02T15:04:05-0700"))
	_, err := conn.Exec(context.Background(), query)
	if err != nil {
		panic(err)
	}
}

func uploadPings(pings []database.PingsRow, conn *pgx.Conn) {
	var query strings.Builder
	query.WriteString("INSERT INTO pings (api_key, url, response_time, status, created_at) VALUES")
	for i, ping := range pings {
		if i > 0 {
			query.WriteString(",")
		}
		query.WriteString(fmt.Sprintf(" ('%s', '%s', %d, %d, '%s')", ping.APIKey, ping.URL, ping.ResponseTime, ping.Status, ping.CreatedAt.UTC().Format("2006-01-02T15:04:05-0700")))
	}
	query.WriteString(";")

	_, err := conn.Exec(context.Background(), query.String())
	if err != nil {
		panic(err)
	}
}

func shuffle(monitored []database.MonitorRow) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(monitored), func(i, j int) {
		monitored[i], monitored[j] = monitored[j], monitored[i]
	})
}

func pingMonitored(monitored []database.MonitorRow) []database.PingsRow {
	client := getClient()

	var pings []database.PingsRow
	for _, m := range monitored {
		status, elapsed, err := ping(client, m.URL, m.Secure, m.Ping)
		if err != nil {
			fmt.Println(err)
		}
		ping := database.PingsRow{
			APIKey:       m.APIKey,
			URL:          m.URL,
			ResponseTime: int(elapsed.Milliseconds()),
			Status:       status,
			CreatedAt:    time.Now(),
		}
		pings = append(pings, ping)
	}
	return pings
}

func getClient() http.Client {
	dialer := net.Dialer{Timeout: 2 * time.Second}
	var client = http.Client{
		Transport: &http.Transport{
			Dial: dialer.Dial,
		},
	}
	return client
}

func main() {
	conn := database.NewConnection()
	defer conn.Close(context.Background())

	monitored := getMonitoredURLs(conn)
	// Shuffle URLs to ping to avoid a page looking consistently slow or fast
	// due to cold starts or caching
	shuffle(monitored)

	pings := pingMonitored(monitored)
	uploadPings(pings, conn)
	deleteExpiredPings(conn)
}
