package testutil

import "time"

type StubTimeProvider struct {
	FixedTime time.Time
}

func (s *StubTimeProvider) Now() time.Time {
	return s.FixedTime
}
