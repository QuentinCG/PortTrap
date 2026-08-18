# Docker usage for PortTrap

This example shows how to build and run the minimal PortTrap image and how to make it work with Fail2Ban on the host.

Build locally:

```sh
docker build -t porttrap:local .
```

Run (recommended): use host networking so PortTrap sees actual incoming connections and bind to the real host ports.

```sh
# create host log dir
sudo mkdir -p /var/log/porttrap
sudo chown $(id -u):$(id -g) /var/log/porttrap

docker run -d \
  --name porttrap \
  --network host \
  -v /var/log/porttrap:/var/log/porttrap \
  -e LOG_FILE=/var/log/porttrap/porttrap.log \
  -e LOG_STDOUT=false \
  porttrap:local
```

Notes:
- Using `--network host` lets PortTrap bind the same ports as the host and avoids port mapping issues. It also makes logs usable by host Fail2Ban by mounting `/var/log/porttrap` into the container.
- If you prefer not to use host networking, map ports explicitly and add `--cap-add=NET_BIND_SERVICE` to allow binding privileged ports (<1024) without running as root.
