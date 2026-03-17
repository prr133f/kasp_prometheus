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

	if isContainerByFiles() {
		return HostTypeContainer
	}

	if virt, err := detectViaOpenRC(); err == nil && virt != HostTypeUnknown {
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

func detectViaOpenRC() (HostType, error) {
	cmd := exec.Command("virt-what")
	if err := cmd.Run(); err != nil {
		return HostTypeUnknown, err
	}

	output, err := cmd.Output()
	if err != nil {
		return HostTypeUnknown, err
	}

	virt := strings.TrimSpace(strings.ToLower(string(output)))

	switch virt {
	case "":
		return HostTypePhysical, nil
	case "docker", "podman", "lxc", "openvz", "rkt":
		return HostTypeContainer, nil
	case "kvm", "vmware", "virtualbox", "xen", "hyperv", "qemu", "bhyve", "parallels":
		return HostTypeVM, nil
	default:
		return HostTypeVM, nil
	}
}

func isContainerByFiles() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	if _, err := os.Stat("/proc/vz"); err == nil {
		return true
	}
	if cgroup, err := os.ReadFile("/proc/self/cgroup"); err == nil {
		if strings.Contains(string(cgroup), "docker") ||
			strings.Contains(string(cgroup), "podman") ||
			strings.Contains(string(cgroup), "lxc") {
			return true
		}
	}
	return false
}

func GetHostMetadata() map[string]string {
	hostname, _ := os.Hostname()

	return map[string]string{
		"hostname": hostname,
		"os":       runtime.GOOS,
		"arch":     runtime.GOARCH,
	}
}
