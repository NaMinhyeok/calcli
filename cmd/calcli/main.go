package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/NaMinhyeok/calcli/internal/app"
	"github.com/NaMinhyeok/calcli/internal/config"
	"github.com/NaMinhyeok/calcli/internal/domain"
	"github.com/NaMinhyeok/calcli/internal/storage/vdir"
	"github.com/NaMinhyeok/calcli/internal/util"
)

func loadConfigAndCalendar() (*config.Config, domain.Calendar) {
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

	return cfg, calendar
}

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
		fmt.Println("calcli v0.0.3")
		os.Exit(0)
	}

	if *help || flag.NArg() == 0 {
		flag.Usage()
		os.Exit(0)
	}

	command := flag.Arg(0)

	switch command {
	case "list":
		listFlags := flag.NewFlagSet("list", flag.ExitOnError)
		fromFlag := listFlags.String("from", "", "Start date (YYYY-MM-DD or 'today')")
		toFlag := listFlags.String("to", "", "End date (YYYY-MM-DD or 'today')")
		listFlags.Parse(flag.Args()[1:])

		var fromTime, toTime *time.Time
		if *fromFlag != "" {
			t, err := util.ParseDate(*fromFlag)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid from date: %v\n", err)
				os.Exit(2)
			}
			fromTime = &t
		}
		if *toFlag != "" {
			t, err := util.ParseDate(*toFlag)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid to date: %v\n", err)
				os.Exit(2)
			}
			toTime = &t
		}

		_, calendar := loadConfigAndCalendar()

		// Use calendar path from config
		reader := vdir.NewReader(os.DirFS(calendar.Path), ".")
		formatter := &app.SimpleEventFormatter{}
		if err := app.ListHandler(reader, formatter, os.Stdout, fromTime, toTime); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "new":
		newFlags := flag.NewFlagSet("new", flag.ExitOnError)
		titleFlag := newFlags.String("title", "New Event", "Event title")
		whenFlag := newFlags.String("when", time.Now().Format("15:04"), "Event start time")
		durationFlag := newFlags.String("duration", "1h", "Event duration")
		locationFlag := newFlags.String("location", "", "Event location")
		newFlags.Parse(flag.Args()[1:])

		title := *titleFlag
		when := *whenFlag
		duration := *durationFlag
		location := *locationFlag

		_, calendar := loadConfigAndCalendar()

		// Create writer and handle new event
		writer := vdir.NewWriter(calendar.Path)
		timeProvider := &util.RealTimeProvider{}
		uidGen := &app.RealUIDGenerator{}
		if err := app.NewHandler(writer, timeProvider, uidGen, title, when, duration, location); err != nil {
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

		_, calendar := loadConfigAndCalendar()

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
		_, calendar := loadConfigAndCalendar()

		// Create importer and UID generator
		writer := vdir.NewWriter(calendar.Path)
		uidGen := &app.RealUIDGenerator{}

		if err := app.ImportHandler(writer, uidGen, filePath, false); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "calendars":
		cfg, _ := loadConfigAndCalendar()

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
