// Package vdir implements vdir-based calendar storage.
// It provides Reader for reading .ics files from vdir directories using io/fs.FS abstraction.
// Future: atomic file operations using os.CreateTemp and os.Rename for writing.
package vdir
