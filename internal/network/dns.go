package network

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const dnsmasqConfigDir = "/etc/dnsmasq.d"
const dnsmasqConfigFile = dnsmasqConfigDir + "/tavpbox.conf"

func SetupDnsmasq(domainSuffix string) error {
	exec.Command("apt-get", "install", "-y", "dnsmasq").Run()

	os.MkdirAll(dnsmasqConfigDir, 0755)

	config := fmt.Sprintf(`# TAVPBox DNS — auto-generated
# Wildcard: all *.tavp.local resolve to localhost
address=/%s/127.0.0.1

# Don't read /etc/resolv.conf (we ARE the resolver)
no-resolv
server=8.8.8.8
server=1.1.1.1

# Listen on localhost
listen-address=127.0.0.1
bind-interfaces

# Cache
cache-size=1000
`, domainSuffix)

	os.WriteFile(dnsmasqConfigFile, []byte(config), 0644)

	resolv := "nameserver 127.0.0.1\n"
	os.WriteFile("/etc/resolv.conf", []byte(resolv), 0644)

	exec.Command("systemctl", "restart", "dnsmasq").Run()
	exec.Command("systemctl", "enable", "dnsmasq").Run()

	return nil
}

func AddDnsmasqEntry(name, ip string) error {
	data, _ := os.ReadFile(dnsmasqConfigFile)
	entry := fmt.Sprintf("address=/%s/%s\n", name, ip)

	if strings.Contains(string(data), name) {
		return nil
	}

	f, _ := os.OpenFile(dnsmasqConfigFile, os.O_APPEND|os.O_WRONLY, 0644)
	f.WriteString(entry)
	f.Close()

	exec.Command("systemctl", "reload", "dnsmasq").Run()
	return nil
}

func RemoveDnsmasqEntry(name string) error {
	data, _ := os.ReadFile(dnsmasqConfigFile)
	lines := strings.Split(string(data), "\n")
	var newLines []string
	for _, line := range lines {
		if !strings.Contains(line, name) || !strings.Contains(line, "address=") {
			newLines = append(newLines, line)
		}
	}
	os.WriteFile(dnsmasqConfigFile, []byte(strings.Join(newLines, "\n")), 0644)
	exec.Command("systemctl", "reload", "dnsmasq").Run()
	return nil
}
