package logging

import (
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "sync"
    "time"

    "porttrap/internal/config"
)

type Event struct {
    Timestamp       time.Time `json:"timestamp"`
    Protocol        string    `json:"protocol"`
    SourceIP        string    `json:"source_ip"`
    SourcePort      int       `json:"source_port"`
    DestinationPort int       `json:"destination_port"`
}

type Logger struct {
    cfg       *config.Config
    ch        chan Event
    wg        sync.WaitGroup
    file      *os.File
    written   int64
    mu        sync.Mutex
    closed    chan struct{}
    dropCount int64
    subs      []chan Event
}

func NewLogger(cfg *config.Config) (*Logger, error) {
    if cfg == nil {
        return nil, errors.New("nil config")
    }
    l := &Logger{
        cfg:    cfg,
        ch:     make(chan Event, 10000),
        closed: make(chan struct{}),
    }

    if cfg.LogFile != "" {
        if err := os.MkdirAll(filepath.Dir(cfg.LogFile), 0o755); err != nil {
            return nil, fmt.Errorf("creating log dir: %w", err)
        }
        f, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
        if err != nil {
            return nil, fmt.Errorf("open log file: %w", err)
        }
        l.file = f
        fi, _ := f.Stat()
        l.written = fi.Size()
    }

    l.wg.Add(1)
    go l.run()
    return l, nil
}

func (l *Logger) Close() {
    close(l.ch)
    l.wg.Wait()
    if l.file != nil {
        l.file.Close()
    }
}

func (l *Logger) Sync() {
    if l.file != nil {
        l.file.Sync()
    }
}

func (l *Logger) run() {
    defer l.wg.Done()
    for ev := range l.ch {
        l.write(ev)
        l.mu.Lock()
        for _, s := range l.subs {
            select {
            case s <- ev:
            default:
            }
        }
        l.mu.Unlock()
    }
}

// Subscribe returns a read-only channel that receives copies of logged events.
// The caller should drain the channel; closing the returned channel is not required.
func (l *Logger) Subscribe() <-chan Event {
    ch := make(chan Event, 1024)
    l.mu.Lock()
    l.subs = append(l.subs, ch)
    l.mu.Unlock()
    return ch
}

func (l *Logger) Log(ev Event) {
    select {
    case l.ch <- ev:
    default:
        // drop if full
        l.mu.Lock()
        l.dropCount++
        l.mu.Unlock()
        fmt.Fprintln(os.Stderr, "logger buffer full: dropping event")
    }
}

func (l *Logger) write(ev Event) {
    b := FormatEvent(ev, l.cfg.LogFormat)

    if l.cfg.LogStdout {
        os.Stdout.Write(b)
    }

    if l.file != nil {
        if _, err := l.file.Write(b); err != nil {
            fmt.Fprintln(os.Stderr, "log write error:", err)
            // attempt to close file to avoid repeated errors
            l.file.Close()
            l.file = nil
            return
        }
        l.written += int64(len(b))
        if l.written >= l.cfg.LogMaxSizeMB*1024*1024 {
            _ = l.rotate()
        }
    }
}

func FormatEvent(ev Event, format string) []byte {
    if format == "json" {
        m := map[string]interface{}{
            "timestamp":        ev.Timestamp.UTC().Format(time.RFC3339),
            "protocol":         ev.Protocol,
            "source_ip":        ev.SourceIP,
            "source_port":      ev.SourcePort,
            "destination_port": ev.DestinationPort,
        }
        jb, _ := json.Marshal(m)
        return append(jb, '\n')
    }
    ts := ev.Timestamp.UTC().Format(time.RFC3339)
    return []byte(fmt.Sprintf("%s %s %s:%d -> %d\n", ts, stringsUpper(ev.Protocol), ev.SourceIP, ev.SourcePort, ev.DestinationPort))
}

func stringsUpper(s string) string {
    // avoid importing strings package twice in small helper
    b := []byte(s)
    for i := range b {
        if b[i] >= 'a' && b[i] <= 'z' {
            b[i] = b[i] - 'a' + 'A'
        }
    }
    return string(b)
}

func (l *Logger) rotate() error {
    l.mu.Lock()
    defer l.mu.Unlock()
    if l.file == nil {
        return errors.New("no file to rotate")
    }
    // close current
    l.file.Close()
    base := l.cfg.LogFile
    // shift files
    for i := l.cfg.LogMaxFiles - 1; i >= 1; i-- {
        src := fmt.Sprintf("%s.%d", base, i)
        dst := fmt.Sprintf("%s.%d", base, i+1)
        if _, err := os.Stat(src); err == nil {
            os.Rename(src, dst)
        }
    }
    // rotate current to .1
    _ = os.Rename(base, base+".1")
    // open new file
    f, err := os.OpenFile(base, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
    if err != nil {
        return err
    }
    l.file = f
    l.written = 0
    return nil
}
