package tmux

import (
	"strconv"
	"strings"
)

// Session represents a tmux session
type Session struct {
	Name     string
	Attached int // Number of attached clients
}

// ParseSessions parses tmux list-sessions output
// Input format: "session_name:attached_count" per line
func ParseSessions(output []string) []Session {
	var sessions []Session

	for _, line := range output {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}

		attached, _ := strconv.Atoi(parts[1])
		sessions = append(sessions, Session{
			Name:     parts[0],
			Attached: attached,
		})
	}

	return sessions
}

// FormatSession returns a display string for a session
func (s Session) FormatSession() string {
	if s.Attached > 0 {
		return s.Name + " (attached)"
	}
	return s.Name
}
