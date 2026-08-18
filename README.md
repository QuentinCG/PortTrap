# PortTrap

PortTrap is a tiny Go program that detects incoming connections on configured ports and logs metadata (timestamp, protocol, source IP/port, destination port). It intentionally does not emulate protocols, respond to clients, or run external commands.

Build:

```sh
go build ./cmd/porttrap
```

Run (example):

```sh
export TCP_PORTS="20,21,22,23,445,1433,3306,3389,5432,5900,6379,8080,27017"
export UDP_PORTS="161,5060"
export LOG_STDOUT=true
./porttrap

Docker

Build a minimal image locally and run it (see `examples/docker/README.md`):

```sh
docker build -t porttrap:local .
sudo mkdir -p /var/log/porttrap && sudo chown $(id -u):$(id -g) /var/log/porttrap
docker run -d --name porttrap --network host -v /var/log/porttrap:/var/log/porttrap -e LOG_FILE=/var/log/porttrap/porttrap.log -e LOG_STDOUT=false porttrap:local
```

Notes:
- Using `--network host` is recommended so PortTrap binds real host ports and Fail2Ban can read the mounted log file.
- To bind privileged ports without host networking, add `--cap-add=NET_BIND_SERVICE` and publish the ports with `-p`.
```

Logs default to text format. Use `LOG_FORMAT=json` for JSON lines.

Fail2Ban

PortTrap writes stable text or JSON log lines. A simple Fail2Ban integration is included under `examples/fail2ban/`.

Files provided:
- `examples/fail2ban/porttrap.conf` — filter matching PortTrap text and JSON logs.
- `examples/fail2ban/jail.local` — example jail configuration (adjust `logpath` if needed).

Installation steps (system-wide Fail2Ban on a Linux server):

1. Ensure PortTrap writes to a file, for example `/var/log/porttrap/porttrap.log`:

```sh
export LOG_FILE=/var/log/porttrap/porttrap.log
export LOG_STDOUT=false
./porttrap &
```

2. Copy filter and jail config into Fail2Ban directories (may require sudo):

```sh
sudo cp examples/fail2ban/porttrap.conf /etc/fail2ban/filter.d/porttrap.conf
sudo cp examples/fail2ban/jail.local /etc/fail2ban/jail.d/porttrap.conf
```

Alternatively, append the `[porttrap]` block into `/etc/fail2ban/jail.local`.

3. Test the filter against your log file to ensure it matches:

```sh
sudo fail2ban-regex /var/log/porttrap/porttrap.log /etc/fail2ban/filter.d/porttrap.conf
```

4. Reload or restart Fail2Ban:

```sh
sudo systemctl restart fail2ban
# or
sudo fail2ban-client reload
```

5. Verify the jail is active:

```sh
sudo fail2ban-client status porttrap
```

Notes:
- `logpath` in `jail.local` must point to the actual PortTrap log file.
- Tweak `maxretry`, `findtime` and `bantime` to suit your policy.
- The included filter supports both text and JSON output — use `LOG_FORMAT` to choose.
- Use `fail2ban-regex` to debug and verify regex matching before enabling bans.
