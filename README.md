# meshify-client

Meshify control plane for WireGuard

## Building Debian package

To build the packages, run `make build`. This will build the necessary Docker
images to build the packages, then build them and place them in the top
directory of the repository.

`make clean` will remove packages, and any Docker containers and images.

