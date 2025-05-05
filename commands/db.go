package commands

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/supabase-community/postgrest-go"
)

func newClient() *postgrest.Client {
	_ = godotenv.Load()
	rawURL := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_KEY")
	url := strings.TrimRight(rawURL, "/")
	if !strings.HasSuffix(url, "/rest/v1") {
		url += "/rest/v1"
	}
	return postgrest.NewClient(
		url,
		"", // public schema
		map[string]string{
			"apikey":        key,
			"Authorization": "Bearer " + key,
		},
	)
}

var supa = newClient()
