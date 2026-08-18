package config

import (
    "errors"
    "fmt"
    "os"
    "sort"
    "strconv"
    "strings"
)

type PortList []int

func (p PortList) String() string {
    parts := make([]string, 0, len(p))
    for _, v := range p {
        parts = append(parts, strconv.Itoa(v))
    }
    return strings.Join(parts, ", ")
}

type Config struct {
    TCPPorts     PortList
    UDPPorts     PortList
    LogStdout    bool
    LogFile      string
    LogMaxSizeMB int64
    LogMaxFiles  int
    LogFormat    string
    MaxConns     int
}

func getenv(key, def string) string {
    v := os.Getenv(key)
    if v == "" {
        return def
    }
    return v
}

func parsePortList(s string) (PortList, error) {
    out := make([]int, 0)
    if strings.TrimSpace(s) == "" {
        return out, nil
    }
    parts := strings.Split(s, ",")
    seen := map[int]struct{}{}
    for i, raw := range parts {
        p := strings.TrimSpace(raw)
        if p == "" {
            return nil, fmt.Errorf("empty port at position %d", i+1)
        }
        v, err := strconv.Atoi(p)
        if err != nil {
            return nil, fmt.Errorf("invalid port '%s'", p)
        }
        if v < 1 || v > 65535 {
            return nil, fmt.Errorf("port out of range: %d", v)
        }
        if _, ok := seen[v]; ok {
            return nil, fmt.Errorf("duplicate port: %d", v)
        }
        seen[v] = struct{}{}
        out = append(out, v)
    }
    sort.Ints(out)
    return out, nil
}

func strconvDefaultInt(env string, def int) (int, error) {
    s := os.Getenv(env)
    if s == "" {
        return def, nil
    }
    v, err := strconv.Atoi(s)
    if err != nil {
        return 0, err
    }
    return v, nil
}

func LoadFromEnv() (*Config, error) {
    tcp := getenv("TCP_PORTS", "")
    udp := getenv("UDP_PORTS", "")
    tcpPorts, err := parsePortList(tcp)
    if err != nil {
        return nil, fmt.Errorf("TCP_PORTS: %w", err)
    }
    udpPorts, err := parsePortList(udp)
    if err != nil {
        return nil, fmt.Errorf("UDP_PORTS: %w", err)
    }

    logStdout := true
    if ls := os.Getenv("LOG_STDOUT"); ls != "" {
        if ls == "false" || ls == "0" {
            logStdout = false
        } else {
            logStdout = true
        }
    }

    logFile := getenv("LOG_FILE", "")

    maxSize := int64(100) // MB
    if s := os.Getenv("LOG_MAX_SIZE_MB"); s != "" {
        v, err := strconv.ParseInt(s, 10, 64)
        if err != nil || v <= 0 {
            return nil, errors.New("LOG_MAX_SIZE_MB must be a positive integer")
        }
        maxSize = v
    }

    maxFiles := 5
    if s := os.Getenv("LOG_MAX_FILES"); s != "" {
        v, err := strconv.Atoi(s)
        if err != nil || v < 1 {
            return nil, errors.New("LOG_MAX_FILES must be a positive integer")
        }
        maxFiles = v
    }

    lf := getenv("LOG_FORMAT", "text")
    if lf != "text" && lf != "json" && lf != "" {
        return nil, errors.New("LOG_FORMAT must be 'text' or 'json'")
    }
    if lf == "" {
        lf = "text"
    }

    maxConns := 1024
    if s := os.Getenv("MAX_CONNECTIONS"); s != "" {
        v, err := strconv.Atoi(s)
        if err != nil || v < 1 {
            return nil, errors.New("MAX_CONNECTIONS must be a positive integer")
        }
        maxConns = v
    }

    if !logStdout && logFile == "" {
        return nil, errors.New("no logging destination enabled: set LOG_STDOUT or LOG_FILE")
    }

    return &Config{
        TCPPorts:     tcpPorts,
        UDPPorts:     udpPorts,
        LogStdout:    logStdout,
        LogFile:      logFile,
        LogMaxSizeMB: maxSize,
        LogMaxFiles:  maxFiles,
        LogFormat:    lf,
        MaxConns:     maxConns,
    }, nil
}
