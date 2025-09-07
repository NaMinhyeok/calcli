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

// exitf prints to stderr and exits with the given code.
func exitf(code int, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(code)
}

// mustParseDatePtr parses a date string or returns nil when empty. Exits on error.
func mustParseDatePtr(v string, name string) *time.Time {
	if v == "" {
		return nil
	}
	t, err := util.ParseDate(v)
	if err != nil {
		exitf(2, "Invalid %s date: %v\n", name, err)
	}
	return &t
}

// helpers to construct storage/format components
func readerFor(calendar domain.Calendar) *vdir.Reader {
	return vdir.NewReader(os.DirFS(calendar.Path), ".")
}

func writerFor(calendar domain.Calendar) *vdir.Writer {
	return vdir.NewWriter(calendar.Path)
}

func formatter(showUID bool) *app.SimpleEventFormatter {
	return &app.SimpleEventFormatter{ShowUID: showUID}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <command> [args]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nCommands:\n")
		fmt.Fprintf(os.Stderr, "  list        List events\n")
		fmt.Fprintf(os.Stderr, "  new         Create new event\n")
		fmt.Fprintf(os.Stderr, "  search      Search events\n")
		fmt.Fprintf(os.Stderr, "  edit        Edit existing event\n")
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
		showUIDFlag := listFlags.Bool("show-uid", false, "Show event UIDs")
		listFlags.Parse(flag.Args()[1:])

		fromTime := mustParseDatePtr(*fromFlag, "from")
		toTime := mustParseDatePtr(*toFlag, "to")

		_, calendar := loadConfigAndCalendar()

		// Use calendar path from config
		reader := readerFor(calendar)
		fmtter := formatter(*showUIDFlag)
		if err := app.ListHandler(reader, fmtter, os.Stdout, fromTime, toTime); err != nil {
			exitf(1, "Error: %v\n", err)
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
		writer := writerFor(calendar)
		timeProvider := &util.RealTimeProvider{}
		uidGen := &app.RealUIDGenerator{}
		if err := app.NewHandler(writer, timeProvider, uidGen, title, when, duration, location); err != nil {
			exitf(1, "Error: %v\n", err)
		}

		fmt.Printf("Event '%s' created successfully\n", title)
	case "search":
		searchFlags := flag.NewFlagSet("search", flag.ExitOnError)
		fieldFlag := searchFlags.String("field", "any", "Field to search in (any, title, desc, location)")
		showUIDFlag := searchFlags.Bool("show-uid", false, "Show event UIDs")
		searchFlags.Parse(flag.Args()[1:])

		if searchFlags.NArg() < 1 {
			exitf(2, "Usage: %s search [--field=any|title|desc|location] <query>\n", os.Args[0])
		}

		query := searchFlags.Arg(0)

		var searchField app.SearchField
		switch *fieldFlag {
		case "any":
			searchField = app.SearchFieldAny
		case "title":
			searchField = app.SearchFieldTitle
		case "desc":
			searchField = app.SearchFieldDesc
		case "location":
			searchField = app.SearchFieldLocation
		default:
			exitf(2, "Invalid field: %s. Valid fields: any, title, desc, location\n", *fieldFlag)
		}

		_, calendar := loadConfigAndCalendar()

		reader := readerFor(calendar)
		fmtter := formatter(*showUIDFlag)
		if err := app.SearchHandler(reader, fmtter, os.Stdout, query, searchField); err != nil {
			exitf(1, "Error: %v\n", err)
		}
	case "edit":
		editFlags := flag.NewFlagSet("edit", flag.ExitOnError)
		uidFlag := editFlags.String("uid", "", "UID of the event to edit (required)")
		titleFlag := editFlags.String("title", "", "New event title")
		whenFlag := editFlags.String("when", "", "New event start time")
		durationFlag := editFlags.String("duration", "", "New event duration")
		locationFlag := editFlags.String("location", "", "New event location")
		editFlags.Parse(flag.Args()[1:])

		if *uidFlag == "" {
			fmt.Fprintf(os.Stderr, "Usage: %s edit --uid=<uid> [--title=<title>] [--when=<when>] [--duration=<duration>] [--location=<location>]\n", os.Args[0])
			exitf(2, "At least one of --title, --when, --duration, or --location must be provided.\n")
		}

		if *titleFlag == "" && *whenFlag == "" && *durationFlag == "" && *locationFlag == "" {
			exitf(2, "At least one edit option must be provided: --title, --when, --duration, or --location\n")
		}

		_, calendar := loadConfigAndCalendar()

		var options app.EditOptions
		if *titleFlag != "" {
			options.Title = titleFlag
		}
		if *whenFlag != "" {
			options.When = whenFlag
		}
		if *durationFlag != "" {
			options.Duration = durationFlag
		}
		if *locationFlag != "" {
			options.Location = locationFlag
		}

		writer := writerFor(calendar)
		timeProvider := &util.RealTimeProvider{}
		if err := app.EditHandler(writer, timeProvider, *uidFlag, options); err != nil {
			exitf(1, "Error: %v\n", err)
		}

		fmt.Printf("Event '%s' updated successfully\n", *uidFlag)
	case "import":
		if flag.NArg() < 2 {
			exitf(2, "Usage: %s import <file.ics>\n", os.Args[0])
		}

		filePath := flag.Arg(1)

		// Load config
		_, calendar := loadConfigAndCalendar()

		// Create importer and UID generator
		writer := writerFor(calendar)
		uidGen := &app.RealUIDGenerator{}

		if err := app.ImportHandler(writer, uidGen, filePath, false); err != nil {
			exitf(1, "Error: %v\n", err)
		}
	case "calendars":
		cfg, _ := loadConfigAndCalendar()

		if err := app.CalendarsHandler(cfg, os.Stdout); err != nil {
			exitf(1, "Error: %v\n", err)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		flag.Usage()
		os.Exit(2)
	}
}
