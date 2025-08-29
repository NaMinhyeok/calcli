# calcli

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**`calcli` is a minimalist, fast, and powerful command-line tool for managing your calendar. If you live in the terminal and prefer simple, text-based tools that are easy to sync and version control, `calcli` is for you.**

---

## The Philosophy: Your Calendar as a `vdir`

`calcli` embraces the [vdir](http://vdirsyncer.pimutils.org/en/stable/vdir.html) format, a simple and effective standard for storing calendar and contact data. This design choice provides several key benefits:

*   **Your Data is Yours:** All events are stored as individual `.ics` files on your local filesystem. There is no opaque database or proprietary format.
*   **Simple & Transparent:** Your calendar is just a directory of files. You can use standard Unix tools like `ls`, `grep`, `cat`, and `git` to interact with your events.
*   **Sync-Friendly:** The vdir format is compatible with tools like [vdirsyncer](https://github.com/pimutils/vdirsyncer) to sync your events with CalDAV servers (e.g., Google Calendar, Nextcloud, Fastmail).
*   **Backup-Friendly:** Backing up your calendar is as simple as copying a directory.

## Features

*   **Intuitive Commands:** A clean, simple API (`new`, `list`, `search`) that is easy to remember and use.
*   **Flexible Time Parsing:** Understands natural time inputs like "14:00" (today at 2 PM) as well as full timestamps.
*   **Multi-Calendar Management:** Keep your `work`, `home`, and `project` calendars cleanly separated in different directories.
*   **Standard `.ics` Format:** Generates RFC 5545 compliant `.ics` files for full compatibility with other calendar applications.
*   **Zero Configuration Required:** Start using it immediately after installation.

## Installation

There are multiple ways to install `calcli`:

#### 1. Via `go install` (Recommended for Go users)

If you have a Go environment set up, this is the simplest method.

```bash
go install github.com/NaMinhyeok/calcli/cmd/calcli@latest
```

**Note:** If you get a "command not found" error after installation, you need to add Go's bin directory to your PATH:

```bash
# For bash users
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# For zsh users (macOS default)
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

#### 2. From GitHub Releases (For non-Go users)

Pre-compiled binaries for various operating systems are available on the [Releases Page](https://github.com/NaMinhyeok/calcli/releases). Download the appropriate archive for your system, extract it, and place the `calcli` binary in a directory included in your system's `PATH`.

## Usage

### `new`: Create a New Event

`calcli new <title> [flags]`

**Examples:**

```bash
# Quick event for today at 2 PM in the 'work' calendar
calcli new "Team Sync" --when "14:00" --calendar work

# A more detailed event with a specific date, duration, and location
calcli new "Project Deadline Planning" --when "2025-10-20 09:00" --duration 3h --location "Main Conference Room" --calendar work

# A personal event in the 'home' calendar
calcli new "Dentist Appointment" --when "2025-10-22 16:30" --duration 45m --calendar home
```

### `list`: List Upcoming Events

`calcli list [flags]`

**Examples:**

```bash
# List all upcoming events from all calendars
calcli list

# List events from the 'work' calendar only
calcli list --calendar work

# List events from multiple specific calendars
calcli list --calendar work --calendar home
```

### `search`: Find Specific Events

`calcli search <keyword>`

Searches for the keyword in the summary (title) of events across all calendars.

**Example:**

```bash
# Find all events related to "Planning"
calcli search "Planning"
```

### `import`: Import an `.ics` File

`calcli import <file_path> --calendar <calendar_name>`

Imports an existing `.ics` file from your local filesystem into a specified calendar.

**Example:**

```bash
# Import a downloaded .ics file into your 'home' calendar
calcli import ~/Downloads/external_event.ics --calendar home
```

### `calendars`: List Your Calendars

`calcli calendars`

Lists all available calendars (i.e., the subdirectories within `~/.calcli/`).

## Configuration

`calcli` is a zero-config tool by default. It stores all data in the `~/.calcli/` directory in your home folder.

To create a new calendar, simply create a new directory within this folder.

```bash
mkdir -p ~/.calcli/project-alpha
```

You can now use `--calendar project-alpha` in your commands.

## Contributing

Contributions are welcome! If you find a bug or have a feature request, please open an issue. If you'd like to contribute code, please open a pull request.

Before submitting a PR, please ensure your code is formatted with `go fmt` and that all tests pass (`go test ./...`). And you can use makefile to run all tests.

## License

This project is licensed under the **MIT License**. See the [LICENSE](https://opensource.org/licenses/MIT) file for details.
