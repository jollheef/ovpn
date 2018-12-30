# ovpn

Simplify VPN configuration management based on easyrsa.

Goals is for omiting as much openvpn settings as possible by default, but then allow to set parameters that user need to change.

## Usage

Create basic configuration for server IP 1.1.1.1 (default port is 443/tcp):

    ovpn --ip 1.1.1.1 init

Create new client configuration:

    ovpn issue client1
