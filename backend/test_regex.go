package main

import (
	"fmt"
	"log"
	"regexp"
)

func parsePostgresURL(dbURL string) (host, port, user, password, dbname string) {
	log.Printf("Input URL: %s", dbURL)
	
	// Parse postgres://user:password@host:port/dbname?params
	re := regexp.MustCompile(`postgres(?:ql)?://([^:]+):([^@]+)@([^:/]+)(?::(\d+))?/([^?]+)`)
	matches := re.FindStringSubmatch(dbURL)
	
	log.Printf("Regex matches count: %d", len(matches))
	for i, m := range matches {
		log.Printf("  Match[%d]: %s", i, m)
	}
	
	if len(matches) >= 6 {
		user = matches[1]
		password = matches[2]
		host = matches[3]
		if matches[4] != "" {
			port = matches[4]
		} else {
			port = "5432"
		}
		dbname = matches[5]
		log.Printf("✅ Parsed successfully")
		log.Printf("   user=%s, password=%s, host=%s, port=%s, dbname=%s", user, password, host, port, dbname)
	} else {
		log.Printf("❌ Failed to parse - not enough matches")
	}
	return
}

func main() {
	testURLs := []string{
		"postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable",
		"postgres://user1:pass1@host1/db1",
		"postgres://user2:pass2@host2:5433/db2",
		"postgresql://user3:pass3@host3/db3?sslmode=disable",
	}
	
	for i, url := range testURLs {
		fmt.Printf("\n════════════════════════════════════════\n")
		fmt.Printf("Test %d:\n", i+1)
		fmt.Printf("════════════════════════════════════════\n")
		parsePostgresURL(url)
	}
}