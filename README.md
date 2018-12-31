# ovpn

Simplify VPN configuration management based on easyrsa.

Goals is for omiting as much openvpn settings as possible by default, but then allow to set parameters that user need to change.

## Usage

Create basic configuration for server IP 192.0.2.53 (default port is 443/tcp):

    ovpn --ip 192.0.2.53 init

Create new client configuration:

    ovpn issue client1

## Server

Copy server configuration:

    [user@localhost:~]$ scp server.nix root@192.0.2.53:/tmp/

After booting from nixos iso:

    [user@localhost:~]$ ssh root@192.0.2.53

    [root@nixos:~]# parted /dev/vda
    (parted) mklabel msdos
    (parted) mkpart primary ext4 0% 100%
    (parted) q
    [root@nixos:~]# mkfs.ext4 -Lroot /dev/vda1
    [root@nixos:~]# mount /dev/vda1 /mnt
    [root@nixos:~]# nixos-generate-config --root /mnt
	[root@nixos:~]# mv /tmp/server.nix /mnt/etc/nixos/configuration.nix
    [root@nixos:~]# nixos-install
    [root@nixos:~]# reboot
