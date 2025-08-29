package main

import (
	"flag"
	"fmt"
	"os"

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
		fmt.Println("new command - not implemented yet")
	case "search":
		fmt.Println("search command - not implemented yet")
	case "import":
		fmt.Println("import command - not implemented yet")
	case "calendars":
		fmt.Println("calendars command - not implemented yet")
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		flag.Usage()
		os.Exit(2)
	}
}
