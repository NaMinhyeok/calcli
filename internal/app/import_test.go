package app

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"calcli/internal/domain"
)

type FakeEventImporter struct {
	events []domain.Event
	err    error
}

func (f *FakeEventImporter) CreateEvent(event domain.Event) error {
	if f.err != nil {
		return f.err
	}
	f.events = append(f.events, event)
	return nil
}

func TestImportHandler(t *testing.T) {
	// Create test ICS file
	tmpDir := t.TempDir()
	icsFile := filepath.Join(tmpDir, "test.ics")

	icsContent := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:test
BEGIN:VEVENT
UID:import-test-1
SUMMARY:Imported Meeting
DTSTART:20250830T140000Z
DTEND:20250830T150000Z
LOCATION:Import Room
DESCRIPTION:Test imported event
END:VEVENT
BEGIN:VEVENT
UID:import-test-2
SUMMARY:Another Import
DTSTART:20250831T100000Z
DTEND:20250831T110000Z
END:VEVENT
END:VCALENDAR`

	err := os.WriteFile(icsFile, []byte(icsContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name          string
		randomUID     bool
		importErr     error
		uidGenErr     error
		expectErr     bool
		expectedCount int
		checkUIDs     bool
	}{
		{
			name:          "basic import",
			randomUID:     false,
			expectedCount: 2,
			checkUIDs:     true,
		},
		{
			name:          "import with random UIDs",
			randomUID:     true,
			expectedCount: 2,
			checkUIDs:     false, // UIDs will be different
		},
		{
			name:      "import error",
			randomUID: false,
			importErr: fmt.Errorf("storage error"),
			expectErr: true,
		},
		{
			name:      "UID generation error with random UID",
			randomUID: true,
			uidGenErr: fmt.Errorf("UID gen error"),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			importer := &FakeEventImporter{err: tt.importErr}
			uidGen := &StubUIDGenerator{uid: "new-random-uid", err: tt.uidGenErr}

			err := ImportHandler(importer, uidGen, icsFile, tt.randomUID)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("expected no error but got: %v", err)
				return
			}

			if len(importer.events) != tt.expectedCount {
				t.Errorf("expected %d events imported, got %d", tt.expectedCount, len(importer.events))
				return
			}

			if tt.checkUIDs {
				expectedUIDs := []string{"import-test-1", "import-test-2"}
				for i, expectedUID := range expectedUIDs {
					if i >= len(importer.events) {
						continue
					}
					if importer.events[i].UID != expectedUID {
						t.Errorf("expected event %d to have UID %q, got %q", i, expectedUID, importer.events[i].UID)
					}
				}
			} else {
				// With random UIDs, check they were changed
				for _, event := range importer.events {
					if event.UID == "import-test-1" || event.UID == "import-test-2" {
						t.Error("expected UID to be changed with random UID option")
					}
				}
			}

			// Check basic event data
			if len(importer.events) > 0 {
				firstEvent := importer.events[0]
				if firstEvent.Summary != "Imported Meeting" {
					t.Errorf("expected first event summary to be 'Imported Meeting', got %q", firstEvent.Summary)
				}
			}
		})
	}
}

func TestImportHandler_FileNotFound(t *testing.T) {
	importer := &FakeEventImporter{}
	uidGen := &StubUIDGenerator{uid: "test-uid"}

	err := ImportHandler(importer, uidGen, "/nonexistent/file.ics", false)

	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
