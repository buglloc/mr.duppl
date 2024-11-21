package dupplcap

import (
	"fmt"
	"strings"

	"go.bug.st/serial/enumerator"
)

const (
	targetVID = "2E8A"
	targetPID = "5052"
)

type Iface struct {
	Path string
	Name string
}

func Ifaces() ([]Iface, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, fmt.Errorf("list detailed ports: %w", err)
	}

	var out []Iface
	for _, port := range ports {
		if !port.IsUSB {
			continue
		}

		if !strings.EqualFold(port.VID, targetVID) {
			continue
		}

		if !strings.EqualFold(port.PID, targetPID) {
			continue
		}

		out = append(out, Iface{
			Path: port.Name,
			Name: fmt.Sprintf("Mr.Duppl:%s", port.SerialNumber),
		})
	}

	return out, nil
}
