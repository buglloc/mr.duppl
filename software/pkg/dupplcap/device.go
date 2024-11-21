package dupplcap

import (
	"fmt"

	"go.bug.st/serial"
)

const (
	cmdStartCapture = 0x01
	cmdStopCapture  = 0x02
)

type Device struct {
	port serial.Port
}

func NewDeviceByIface(iface string) (*Device, error) {
	serialPort, err := serial.Open(iface, &serial.Mode{
		BaudRate: 115200,
	})
	if err != nil {
		return nil, fmt.Errorf("open serial port %s: %w", iface, err)
	}

	return &Device{
		port: serialPort,
	}, nil
}

func NewDeviceByName(name string) (*Device, error) {
	ifaces, err := Ifaces()
	if err != nil {
		return nil, fmt.Errorf("list ifaces: %w", err)
	}

	for _, iface := range ifaces {
		if iface.Name != name {
			continue
		}

		return NewDeviceByIface(iface.Path)
	}

	return nil, fmt.Errorf("device with name %s not found", name)
}

func (r *Device) Packet() ([]byte, error) {
	return readSlipPacket(r.port)
}

func (r *Device) StartCapture(withPacketFolding bool) error {
	cmd := []byte{cmdStartCapture, 0x00}
	if withPacketFolding {
		cmd[1] = 0x01
	}

	return writeSlipPacket(r.port, cmd)
}

func (r *Device) StopCapture() error {
	return writeSlipPacket(r.port, []byte{cmdStopCapture})
}

func (r *Device) Close() error {
	return r.port.Close()
}
