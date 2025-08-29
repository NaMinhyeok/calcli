package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"calcli/internal/app"
	"calcli/internal/config"
	"calcli/internal/storage/vdir"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <command> [args]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nCommands:\n")
		fmt.Fprintf(os.Stderr, "  list        List events\n")
		fmt.Fprintf(os.Stderr, "  new         Create new event\n")
		fmt.Fprintf(os.Stderr, "  search      Search events\n")
		fmt.Fprintf(os.Stderr, "  import      Import events from ICS file\n")
		fmt.Fprintf(os.Stderr, "  calendars   Print available calendars\n")
		fmt.Fprintf(os.Stderr, "\nGlobal flags:\n")
		flag.PrintDefaults()
	}

	version := flag.Bool("version", false, "Show version")
	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	if *version {
		fmt.Println("calcli v0.1.0")
		os.Exit(0)
	}

	if *help || flag.NArg() == 0 {
		flag.Usage()
		os.Exit(0)
	}

	command := flag.Arg(0)

	switch command {
	case "list":
		// Load configuration
		cfg, err := config.Load(config.GetDefaultConfigPath())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
			os.Exit(1)
		}

		// Get default calendar
		calendar, err := cfg.GetDefaultCalendar()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Calendar error: %v\n", err)
			os.Exit(1)
		}

		// Use calendar path from config
		reader := vdir.NewReader(os.DirFS(calendar.Path), ".")
		formatter := &app.SimpleEventFormatter{}
		if err := app.ListHandler(reader, formatter, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "new":
		// Parse flags for new command
		title := "New Event"
		when := time.Now().Format("15:04")
		duration := "1h"

		if flag.NArg() > 1 {
			title = flag.Arg(1)
		}
		if flag.NArg() > 2 {
			when = flag.Arg(2)
		}
		if flag.NArg() > 3 {
			duration = flag.Arg(3)
		}

		// Load config to get calendar path
		cfg, err := config.Load(config.GetDefaultConfigPath())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
			os.Exit(1)
		}

		calendar, err := cfg.GetDefaultCalendar()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Calendar error: %v\n", err)
			os.Exit(1)
		}

		// Create writer and handle new event
		writer := vdir.NewWriter(calendar.Path)
		timeProvider := &app.RealTimeProvider{}
		uidGen := &app.RealUIDGenerator{}
		if err := app.NewHandler(writer, timeProvider, uidGen, title, when, duration); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Event '%s' created successfully\n", title)
	case "search":
		if flag.NArg() < 2 {
			fmt.Fprintf(os.Stderr, "Usage: %s search <query>\n", os.Args[0])
			os.Exit(2)
		}

		query := flag.Arg(1)

		cfg, err := config.Load(config.GetDefaultConfigPath())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
			os.Exit(1)
		}

		calendar, err := cfg.GetDefaultCalendar()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Calendar error: %v\n", err)
			os.Exit(1)
		}

		reader := vdir.NewReader(os.DirFS(calendar.Path), ".")
		formatter := &app.SimpleEventFormatter{}
		if err := app.SearchHandler(reader, formatter, os.Stdout, query, app.SearchFieldAny); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "import":
		if flag.NArg() < 2 {
			fmt.Fprintf(os.Stderr, "Usage: %s import <file.ics>\n", os.Args[0])
			os.Exit(2)
		}

		filePath := flag.Arg(1)

		// Load config
		cfg, err := config.Load(config.GetDefaultConfigPath())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
			os.Exit(1)
		}

		calendar, err := cfg.GetDefaultCalendar()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Calendar error: %v\n", err)
			os.Exit(1)
		}

		// Create importer and UID generator
		writer := vdir.NewWriter(calendar.Path)
		uidGen := &app.RealUIDGenerator{}

		if err := app.ImportHandler(writer, uidGen, filePath, false); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "calendars":
		cfg, err := config.Load(config.GetDefaultConfigPath())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
			os.Exit(1)
		}

		if err := app.CalendarsHandler(cfg, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		flag.Usage()
		os.Exit(2)
	}
}
