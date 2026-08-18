# PortTrap

PortTrap is a tiny Go program that detects incoming connections on configured ports and logs metadata (timestamp, protocol, source IP/port, destination port). It intentionally does not emulate protocols, respond to clients, or run external commands.

## Docker (recommended)

Use the ready-to-run image:

`ghcr.io/quentincg/porttrap:latest`

Run example:

```sh
sudo mkdir -p /var/log/porttrap
sudo chown $(id -u):$(id -g) /var/log/porttrap

docker run -d \
	--name porttrap \
	--network host \
	-v /var/log/porttrap:/var/log/porttrap \
	-e TCP_PORTS="20,21,22,23,445,1433,3306,3389,5432,5900,6379,8080,27017" \
	-e UDP_PORTS="161,5060" \
	-e LOG_FILE=/var/log/porttrap/porttrap.log \
	-e LOG_STDOUT=true \
	-e LOG_FORMAT=text \
	ghcr.io/quentincg/porttrap:latest
```

Notes:
- `--network host` is recommended so PortTrap binds host ports directly.
- Mounting `/var/log/porttrap` lets host Fail2Ban read PortTrap logs.
- If you do not use host networking, publish ports manually and add `--cap-add=NET_BIND_SERVICE` for privileged ports.

More Docker usage details:
- `examples/docker/README.md`

## Build From Source

```sh
go build ./cmd/porttrap
```

Run example:

```sh
export TCP_PORTS="20,21,22,23,445,1433,3306,3389,5432,5900,6379,8080,27017"
export UDP_PORTS="161,5060"
export LOG_STDOUT=true
./porttrap
```

Logs default to text format. Use `LOG_FORMAT=json` for JSON lines.

## Fail2Ban

PortTrap writes stable text or JSON log lines. Ready-to-use Fail2Ban examples are available here:

https://github.com/QuentinCG/PortTrap/tree/main/examples/fail2ban

Files provided:
- `examples/fail2ban/porttrap.conf` — filter matching PortTrap text and JSON logs.
- `examples/fail2ban/jail.local` — example jail configuration (adjust `logpath` if needed).

Installation steps (system-wide Fail2Ban on Linux):

1. Ensure PortTrap writes to `/var/log/porttrap/porttrap.log`.
2. Copy files:

```sh
sudo cp examples/fail2ban/porttrap.conf /etc/fail2ban/filter.d/porttrap.conf
sudo cp examples/fail2ban/jail.local /etc/fail2ban/jail.d/porttrap.conf
```

3. Test regex:

```sh
sudo fail2ban-regex /var/log/porttrap/porttrap.log /etc/fail2ban/filter.d/porttrap.conf
```

4. Reload Fail2Ban:

```sh
sudo systemctl restart fail2ban
```

5. Verify:

```sh
sudo fail2ban-client status porttrap
```
