package logging

import (
    "encoding/json"
    "strings"
    "testing"
    "time"
)

func TestFormatEventText(t *testing.T) {
    ev := Event{Timestamp: time.Date(2026, 8, 18, 9, 42, 10, 0, time.UTC), Protocol: "tcp", SourceIP: "1.2.3.4", SourcePort: 1234, DestinationPort: 6379}
    s := string(FormatEvent(ev, "text"))
    if !strings.Contains(s, "TCP") || !strings.Contains(s, "1.2.3.4:1234") {
        t.Fatalf("unexpected text output: %s", s)
    }
}

func TestFormatEventJSON(t *testing.T) {
    ev := Event{Timestamp: time.Date(2026, 8, 18, 9, 42, 10, 0, time.UTC), Protocol: "udp", SourceIP: "::1", SourcePort: 53, DestinationPort: 161}
    b := FormatEvent(ev, "json")
    var m map[string]interface{}
    if err := json.Unmarshal(b, &m); err != nil {
        t.Fatalf("json unmarshal: %v", err)
    }
    if m["protocol"] != "udp" {
        t.Fatalf("unexpected protocol: %v", m["protocol"])
    }
}
