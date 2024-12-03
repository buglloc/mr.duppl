package dupplcap

import (
	"fmt"
	"strings"

	"go.bug.st/serial/enumerator"
)

const (
	VID  = "2E8A"
	PID  = "5052"
	Name = "Mr.Duppl"
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

		if !strings.EqualFold(port.VID, VID) {
			continue
		}

		if !strings.EqualFold(port.PID, PID) {
			continue
		}

		out = append(out, Iface{
			Path: port.Name,
			Name: fmt.Sprintf("%s:%s", Name, port.SerialNumber),
		})
	}

	return out, nil
}
