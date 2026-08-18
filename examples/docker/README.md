# Docker usage for PortTrap

This example shows how to run the published PortTrap image and how to make it work with Fail2Ban on the host.

Published image:

`ghcr.io/quentincg/porttrap:latest`

Run (recommended): use host networking so PortTrap sees actual incoming connections and bind to the real host ports.

```sh
# create host log dir
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

Build locally (optional):

```sh
docker build -t porttrap:local .
```

Notes:
- Using `--network host` lets PortTrap bind the same ports as the host and avoids port mapping issues. It also makes logs usable by host Fail2Ban by mounting `/var/log/porttrap` into the container.
- If you prefer not to use host networking, map ports explicitly and add `--cap-add=NET_BIND_SERVICE` to allow binding privileged ports (<1024) without running as root.
