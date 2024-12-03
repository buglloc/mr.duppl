package dupplcap

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"go.bug.st/serial"
)

const (
	cmdStartCapture = 0x01
	cmdStopCapture  = 0x02
)

type Device struct {
	iface string
	port  serial.Port
}

func NewDevice(nameOrIface string) (*Device, error) {
	if len(nameOrIface) != 0 {
		if strings.HasPrefix(nameOrIface, Name) {
			return NewDeviceByName(nameOrIface)
		}

		return NewDeviceByIface(nameOrIface)
	}

	ifaces, err := Ifaces()
	if err != nil {
		return nil, fmt.Errorf("list ifaces: %w", err)
	}

	if len(ifaces) == 0 {
		return nil, errors.New("device was not found")
	}

	return NewDeviceByIface(ifaces[0].Path)
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

func NewDeviceByIface(iface string) (*Device, error) {
	serialPort, err := serial.Open(iface, &serial.Mode{
		BaudRate: 115200,
	})
	if err != nil {
		return nil, fmt.Errorf("open serial port %s: %w", iface, err)
	}

	return &Device{
		iface: iface,
		port:  serialPort,
	}, nil
}

func (r *Device) Iface() string {
	return r.iface
}

func (r *Device) Packet() ([]byte, error) {
	packet, err := readSlipPacket(r.port)
	if err != nil {
		var portErr *serial.PortError
		if errors.As(err, &portErr) && portErr.Code() == serial.PortClosed {
			return nil, io.EOF
		}

		return nil, fmt.Errorf("read packet: %w", err)
	}

	return packet, nil
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
