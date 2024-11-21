package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
	"github.com/gopacket/gopacket/pcapgo"
	"github.com/kor44/extcap"

	"github.com/buglloc/Mr.Duppl/software/pkg/dupplcap"
)

const (
	usePacketFoldingOpt = "packet-folding"
)

var (
	usePacketFolding = extcap.NewConfigBoolOpt(usePacketFoldingOpt, "Use packet folding").
		Default(true)
)

func main() {
	app := extcap.App{
		Usage: "dupplcap",
		Version: extcap.VersionInfo{
			Info: "1.0.0",
			Help: "https://github.com/buglloc/Mr.Duppl",
		},
		HelpPage:            "DupplCAP - extcap application to integrate Mr.Duppl with Wireshark or something",
		GetInterfaces:       getAllInterfaces,
		GetDLT:              getDLT,
		GetAllConfigOptions: getAllConfigOptions,
		GetConfigOptions:    getConfigOptions,
		StartCapture:        startCapture,
	}

	app.Run(os.Args)
}

func getAllInterfaces() ([]extcap.CaptureInterface, error) {
	ifaces, err := dupplcap.Ifaces()
	if err != nil {
		return nil, fmt.Errorf("unable to get information about interfaces: %w", err)
	}

	extIfaces := make([]extcap.CaptureInterface, len(ifaces))
	for i, iface := range ifaces {
		// we use Name as Value to deal with Mr.Duppl replugs
		extIfaces[i] = extcap.CaptureInterface{
			Value:   iface.Name,
			Display: iface.Path,
		}
	}

	return extIfaces, nil
}

func getDLT(_ string) (extcap.DLT, error) {
	return extcap.DLT{
		Number: dupplcap.LinkTypeUSBFullSpeed.Int(),
		Name:   dupplcap.LinkTypeUSBFullSpeed.String(),
	}, nil
}

func getConfigOptions(_ string) ([]extcap.ConfigOption, error) {
	opts := []extcap.ConfigOption{
		usePacketFolding,
	}

	return opts, nil
}

func getAllConfigOptions() []extcap.ConfigOption {
	opts := []extcap.ConfigOption{
		usePacketFolding,
	}
	return opts
}

func startCapture(iface string, pipe io.WriteCloser, _ string, opts map[string]any) error {
	defer func() { _ = pipe.Close() }()

	dev, err := dupplcap.NewDeviceByName(iface)
	if err != nil {
		return fmt.Errorf("open device: %w", err)
	}
	defer func() { _ = dev.Close() }()

	packetFoldingEnabled := true
	if val, ok := opts[usePacketFoldingOpt]; ok {
		if en, ok := val.(bool); ok {
			packetFoldingEnabled = en
		}
	}

	if err := dev.StartCapture(packetFoldingEnabled); err != nil {
		return fmt.Errorf("start capture: %w", err)
	}
	defer func() { _ = dev.StopCapture() }()

	w, err := pcapgo.NewNgWriterInterface(
		pipe,
		pcapgo.NgInterface{
			Name:       filepath.Base(iface),
			OS:         runtime.GOOS,
			LinkType:   layers.LinkType(dupplcap.LinkTypeUSBFullSpeed.Int()),
			SnapLength: 0, //unlimited
			// TimestampResolution: 9,
		},
		pcapgo.NgWriterOptions{
			SectionInfo: pcapgo.NgSectionInfo{
				Hardware:    runtime.GOARCH,
				OS:          runtime.GOOS,
				Application: "Mr.Duppl", //spread the word
			},
		},
	)
	if err != nil {
		return fmt.Errorf("open pcapng writer: %w", err)
	}

	for {
		packet, err := dev.Packet()
		if err != nil {
			return fmt.Errorf("read packet: %w", err)
		}

		ci := gopacket.CaptureInfo{
			Length:         len(packet),
			CaptureLength:  len(packet),
			InterfaceIndex: 0,
		}
		err = w.WritePacket(ci, packet)
		if err != nil {
			return fmt.Errorf("write packet: %w", err)
		}

		err = w.Flush()
		if err != nil {
			return fmt.Errorf("flush packet: %w", err)
		}
	}
}
