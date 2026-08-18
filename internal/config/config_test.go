package config

import "testing"

func TestParsePortListValid(t *testing.T) {
    inp := "22, 23,6379"
    p, err := parsePortList(inp)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(p) != 3 {
        t.Fatalf("expected 3 ports, got %d", len(p))
    }
}

func TestParsePortListInvalid(t *testing.T) {
    cases := []string{"22,abc,6379", "0,23", "65536"}
    for _, c := range cases {
        if _, err := parsePortList(c); err == nil {
            t.Fatalf("expected error for input %q", c)
        }
    }
}

func TestParsePortListDuplicate(t *testing.T) {
    if _, err := parsePortList("22,22"); err == nil {
        t.Fatal("expected duplicate port error")
    }
}
