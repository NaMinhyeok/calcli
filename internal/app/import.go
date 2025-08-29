package app

import (
	"fmt"
	"os"

	"calcli/internal/domain"
	"calcli/internal/ical"
)

type EventImporter interface {
	CreateEvent(event domain.Event) error
}

func ImportHandler(importer EventImporter, uidGen UIDGenerator, filePath string, randomUID bool) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	events, err := ical.ParseEvents(file)
	if err != nil {
		return fmt.Errorf("failed to parse ICS file: %v", err)
	}

	imported := 0
	for _, event := range events {
		if randomUID {
			newUID, err := uidGen.Generate()
			if err != nil {
				return fmt.Errorf("failed to generate UID for event %s: %v", event.UID, err)
			}
			event.UID = newUID
		}

		if err := importer.CreateEvent(event); err != nil {
			return fmt.Errorf("failed to import event %s: %v", event.UID, err)
		}
		imported++
	}

	fmt.Printf("Successfully imported %d events\n", imported)
	return nil
}
