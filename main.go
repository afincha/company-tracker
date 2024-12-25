package main

import (
	"log"
	"os"
	"strings"

	"companies-test/firecrawl"
	_ "companies-test/migrations"

	"github.com/joho/godotenv"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	app := pocketbase.New()
	firecrawlClient := firecrawl.NewClient()

	// loosely check if it was executed using "go run"
	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		// enable auto creation of migration files when making collection changes in the Dashboard
		// (the isGoRun check is to enable it only during development)
		Automigrate: isGoRun,
	})

	app.OnRecordAfterCreateSuccess("companies").BindFunc(func(e *core.RecordEvent) error {
		website := e.Record.GetString("website")

		// Scrape website
		resp, err := firecrawlClient.ScrapeURL(website)
		if err != nil {
			log.Fatalf("Failed to scrape website: %v", err)
			return e.Next()
		}

		e.Record.Set("description", resp.Data.Extract.CompanyDescription)
		e.Record.Set("product_summary", resp.Data.Extract.ProductSummary)
		e.App.Save(e.Record)

		return e.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
