package server

import (
    "context"
    "errors"
    "fmt"
    "net"
    "os"
    "sync"
    "syscall"
    "time"

    "porttrap/internal/config"
    "porttrap/internal/listener"
    "porttrap/internal/logging"
)

type Server struct {
    cfg    *config.Config
    logger *logging.Logger
    lis    []net.Listener
    cancel context.CancelFunc
}

func New(cfg *config.Config, logger *logging.Logger) *Server {
    return &Server{cfg: cfg, logger: logger}
}

func listenTCP(port int, network string) (net.Listener, error) {
    addr := fmt.Sprintf(":%d", port)
    return net.Listen(network, addr)
}

func StartListeners(ctx context.Context, cfg *config.Config, logger *logging.Logger, wg *sync.WaitGroup) ([]net.Listener, error) {
    var listeners []net.Listener
    sem := make(chan struct{}, cfg.MaxConns)
    // TCP
    for _, p := range cfg.TCPPorts {
        boundAny := false
        // try tcp4 and tcp6
        for _, netw := range []string{"tcp4", "tcp6"} {
            l, err := listenTCP(p, netw)
            if err != nil {
                // report bind failure for this address family with a helpful hint
                var reason string
                switch {
                case errors.Is(err, syscall.EACCES), errors.Is(err, syscall.EPERM):
                    reason = "permission denied (insufficient privileges)"
                case errors.Is(err, syscall.EADDRINUSE):
                    reason = "port already in use"
                default:
                    reason = err.Error()
                }
                fmt.Fprintf(os.Stderr, "failed to bind %s :%d: %s\n", netw, p, reason)
                continue
            }
            boundAny = true
            listeners = append(listeners, l)
            wg.Add(1)
            // capture port and network values into closure parameters to avoid loop-variable capture
            go func(li net.Listener, port int, network string) {
                defer wg.Done()
                // do not log a synthetic startup event here; only real connections are logged by the listener
                listener.ServeTCP(ctx, li, network, logger, sem, nil)
            }(l, p, netw)
            fmt.Printf("TCP listener started on :%d (%s)\n", p, netw)
        }
        if !boundAny {
            fmt.Fprintf(os.Stderr, "warning: no TCP listener bound for port %d (tcp4 and tcp6 failed)\n", p)
        }
    }

    // UDP
    for _, p := range cfg.UDPPorts {
        boundAny := false
        for _, netw := range []string{"udp4", "udp6"} {
            addr := fmt.Sprintf(":%d", p)
            udpAddr, err := net.ResolveUDPAddr(netw, addr)
            if err != nil {
                continue
            }
            conn, err := net.ListenUDP(netw, udpAddr)
            if err != nil {
                var reason string
                switch {
                case errors.Is(err, syscall.EACCES), errors.Is(err, syscall.EPERM):
                    reason = "permission denied (insufficient privileges)"
                case errors.Is(err, syscall.EADDRINUSE):
                    reason = "address already in use"
                default:
                    reason = err.Error()
                }
                fmt.Fprintf(os.Stderr, "failed to bind %s :%d: %s\n", netw, p, reason)
                continue
            }
            boundAny = true
            wg.Add(1)
            // capture port and network to avoid loop-variable capture
            go func(c *net.UDPConn, port int, network string) {
                defer wg.Done()
                fmt.Printf("UDP listener started on :%d (%s)\n", port, network)
                listener.ServeUDP(ctx, c, logger)
            }(conn, p, netw)
        }
        if !boundAny {
            fmt.Fprintf(os.Stderr, "warning: no UDP listener bound for port %d (udp4 and udp6 failed)\n", p)
        }
    }

    return listeners, nil
}

func (s *Server) Start(ctx context.Context) error {
    ctx, cancel := context.WithCancel(ctx)
    s.cancel = cancel
    var wg sync.WaitGroup
    _, err := StartListeners(ctx, s.cfg, s.logger, &wg)
    if err != nil {
        return err
    }
    go func() {
        wg.Wait()
        cancel()
    }()
    return nil
}

func (s *Server) Stop(ctx context.Context) {
    if s.cancel != nil {
        s.cancel()
    }
    // nothing else since listeners close on context cancellation
}

func nowUTC() (t time.Time) { return time.Now().UTC() }
