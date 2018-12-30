package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func checkUtils(utils ...string) (err error) {
	// Check for required commands
	for _, cmd := range utils {
		_, err := exec.Command("which", cmd).CombinedOutput()
		if err != nil {
			return fmt.Errorf("Command not found: %s", cmd)
		}
	}
	return
}

func run(name string, arg ...string) {
	log.Println(name, arg)
	output, err := exec.Command(name, arg...).CombinedOutput()
	if err != nil {
		log.Println(string(output))
		log.Fatalln(err)
	}
}

func initialize(iface, ip, port, net, mask string) {
	run("easyrsa-init")
	run("easyrsa", "init-pki")
	run("easyrsa", "build-ca", "nopass")
	run("easyrsa", "gen-dh")
	run("openvpn", "--genkey", "--secret", "pki/static.key")

	name := "server"
	run("easyrsa", "build-server-full", name, "nopass")

	s := fmt.Sprintf("port %s\n", port)
	s += "proto tcp\n"
	s += fmt.Sprintf("dev %s\n", iface)
	s += "dev-type tun\n"
	s += fmt.Sprintf("server %s %s\n", net, mask)
	s += "push \"redirect-gateway def1\"\n"
	s += "keepalive 10 120\n"
	s += "comp-lzo\n"
	s += "persist-key\n"
	s += "persist-tun\n"
	s += "user nobody\n"
	s += "group nogroup\n"
	s += "tls-server\n"
	s += "key-direction 0\n"

	s += "<ca>\n"
	crt, err := ioutil.ReadFile("pki/ca.crt")
	if err != nil {
		log.Fatalln(err)
	}
	s += string(crt)
	s += "</ca>\n"

	s += "<cert>\n"
	cert, err := ioutil.ReadFile(fmt.Sprintf("pki/issued/%s.crt", name))
	if err != nil {
		log.Fatalln(err)
	}
	s += string(cert)
	s += "</cert>\n"

	s += "<dh>\n"
	dh, err := ioutil.ReadFile("pki/dh.pem")
	if err != nil {
		log.Fatalln(err)
	}
	s += string(dh)
	s += "</dh>\n"

	s += "<key>\n"
	key, err := ioutil.ReadFile(fmt.Sprintf("pki/private/%s.key", name))
	if err != nil {
		log.Fatalln(err)
	}
	s += string(key)
	s += "</key>\n"

	s += "<tls-auth>\n"
	tls, err := ioutil.ReadFile("pki/static.key")
	if err != nil {
		log.Fatalln(err)
	}
	s += string(tls)
	s += "</tls-auth>\n"

	err = ioutil.WriteFile("ovpn/server.ovpn", []byte(s), 0644)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Server configuration file is saved to ovpn/server.ovpn")

	// TODO generate nixos configuration
}

func issue(name, iface, ip, port string) {
	run("easyrsa", "build-client-full", name, "nopass")

	// FIXME a lot of copy paste between this and `initialize`

	s := "client\n"
	s += "proto tcp\n"
	s += fmt.Sprintf("dev %s\n", iface)
	s += "dev-type tun\n"
	s += fmt.Sprintf("remote %s %s\n", ip, port)
	s += "resolv-retry infinite\n"
	s += "nobind\n"
	s += "persist-key\n"
	s += "persist-tun\n"
	s += "comp-lzo\n"
	s += "remote-cert-tls server\n"
	s += "tls-client\n"
	s += "key-direction 1\n"

	s += "<ca>\n"
	crt, err := ioutil.ReadFile("pki/ca.crt")
	if err != nil {
		log.Fatalln(err)
	}
	s += string(crt)
	s += "</ca>\n"

	s += "<cert>\n"
	cert, err := ioutil.ReadFile(fmt.Sprintf("pki/issued/%s.crt", name))
	if err != nil {
		log.Fatalln(err)
	}
	s += string(cert)
	s += "</cert>\n"

	s += "<key>\n"
	key, err := ioutil.ReadFile(fmt.Sprintf("pki/private/%s.key", name))
	if err != nil {
		log.Fatalln(err)
	}
	s += string(key)
	s += "</key>\n"

	s += "<tls-auth>\n"
	tls, err := ioutil.ReadFile("pki/static.key")
	if err != nil {
		log.Fatalln(err)
	}
	s += string(tls)
	s += "</tls-auth>\n"

	err = ioutil.WriteFile(fmt.Sprintf("ovpn/%s.ovpn", name), []byte(s), 0644)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(fmt.Sprintf("Client configuration file is saved to ovpn/%s.ovpn", name))
}

func main() {
	err := checkUtils("easyrsa", "easyrsa-init", "openvpn")
	if err != nil {
		log.Fatalln(err)
	}

	os.Mkdir("ovpn", 0755)

	iface := kingpin.Flag("iface", "Interface name").Default("vpn").String()
	port := kingpin.Flag("port", "Server port").Default("443").String()
	net := kingpin.Flag("net", "Network").Default("10.0.0.0").String()
	mask := kingpin.Flag("mask", "Network mask").Default("255.255.255.0").String()

	initCmd := kingpin.Command("init", "Produce a new server config")

	issueCmd := kingpin.Command("issue", "Produce a new client config")
	issueName := issueCmd.Arg("name", "Name of client").Required().String()

	// TODO do something like that with iface/port/net/mask/...
	ipFlag := kingpin.Flag("ip", "Server IP address")
	rawIp, err := ioutil.ReadFile("ip")
	var ip *string
	if err == nil {
		s := strings.TrimSpace(string(rawIp))
		ip = ipFlag.Default(s).String()
	} else {
		ip = ipFlag.Required().String()
		kingpin.Parse()

		err = ioutil.WriteFile("ip", []byte(*ip), 0644)
		if err != nil {
			log.Fatalln(err)
		}
	}

	os.Setenv("EASYRSA_BATCH", "1")

	switch kingpin.Parse() {
	case initCmd.FullCommand():
		initialize(*iface, *ip, *port, *net, *mask)
	case issueCmd.FullCommand():
		issue(*issueName, *iface, *ip, *port)
	}
}
