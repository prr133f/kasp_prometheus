package detect

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type HostType string

const (
	HostTypeVM        HostType = "vm"
	HostTypeContainer HostType = "container"
	HostTypePhysical  HostType = "physical"
	HostTypeUnknown   HostType = "unknown"
)

func DetectVirtualization() HostType {
	if virt, err := detectViaSystemd(); err == nil && virt != HostTypeUnknown {
		return virt
	}

	return HostTypePhysical
}

func detectViaSystemd() (HostType, error) {
	cmd := exec.Command("systemd-detect-virt")
	output, err := cmd.Output()
	if err != nil {
		return HostTypeUnknown, err
	}

	virt := strings.TrimSpace(strings.ToLower(string(output)))

	switch virt {
	case "", "none":
		return HostTypePhysical, nil
	case "docker", "podman", "lxc", "openvz", "rkt":
		return HostTypeContainer, nil
	case "kvm", "vmware", "virtualbox", "xen", "hyperv", "qemu", "bhyve", "parallels":
		return HostTypeVM, nil
	default:
		return HostTypeVM, nil
	}
}

func GetHostMetadata() map[string]string {
	hostname, _ := os.Hostname()

	return map[string]string{
		"hostname": hostname,
		"os":       runtime.GOOS,
		"arch":     runtime.GOARCH,
	}
}
