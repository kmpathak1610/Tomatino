package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://foodapp:foodapp@localhost:5432/fooddelivery?sslmode=disable"
	}
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		fmt.Println("connect error:", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	dir := "migrations"
	entries, _ := os.ReadDir(dir)
	var ups []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			ups = append(ups, e.Name())
		}
	}
	sort.Strings(ups)
	for _, f := range ups {
		b, err := os.ReadFile(filepath.Join(dir, f))
		if err != nil {
			fmt.Println("read error:", err)
			os.Exit(1)
		}
		_, err = conn.Exec(context.Background(), string(b))
		if err != nil {
			fmt.Println("exec error on", f, ":", err)
			os.Exit(1)
		}
		fmt.Println("applied", f)
	}
	fmt.Println("done")
	_ = bufio.NewReader
}