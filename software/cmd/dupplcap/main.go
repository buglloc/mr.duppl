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

	"github.com/buglloc/mr.duppl/software/pkg/dupplcap"
)

const (
	usePacketFoldingOpt = "packet-folding"
)

func newUsePacketFoldingArg() *extcap.BoolArg {
	return &extcap.BoolArg{
		Name:    usePacketFoldingOpt,
		Display: "Use packet folding",
		Default: true,
	}
}

func main() {
	app := extcap.App{
		Description:    "DupplCAP - extcap application to integrate Mr.Duppl with Wireshark or something",
		Version:        "1.0.0",
		HelpPage:       "https://github.com/buglloc/mr.duppl",
		ListInterfaces: getAllInterfaces,
		ListLinkType:   getDLT,
		ListArgs:       getConfigOptions,
		StartCapture:   startCapture,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func getAllInterfaces() ([]extcap.Interface, error) {
	ifaces, err := dupplcap.Ifaces()
	if err != nil {
		return nil, fmt.Errorf("unable to get information about interfaces: %w", err)
	}

	extIfaces := make([]extcap.Interface, len(ifaces))
	for i, iface := range ifaces {
		// we use Name as Value to deal with Mr.Duppl replugs
		extIfaces[i] = extcap.Interface{
			Value:   iface.Name,
			Display: iface.Path,
		}
	}

	return extIfaces, nil
}

func getDLT(_ string) (extcap.LinkType, error) {
	return extcap.LinkType{
		Number: dupplcap.LinkTypeUSBFullSpeed.Int(),
		Name:   dupplcap.LinkTypeUSBFullSpeed.String(),
	}, nil
}

func getConfigOptions(_ string) ([]extcap.Arg, error) {
	return []extcap.Arg{
		newUsePacketFoldingArg(),
	}, nil
}

func startCapture(iface string, pipe io.WriteCloser, _ string, args *extcap.CaptureArgs) error {
	defer func() { _ = pipe.Close() }()

	dev, err := dupplcap.NewDeviceByName(iface)
	if err != nil {
		return fmt.Errorf("open device: %w", err)
	}
	defer func() { _ = dev.Close() }()

	usePacketFolding := newUsePacketFoldingArg()
	if err := args.ParseArgs([]extcap.Arg{usePacketFolding}); err != nil {
		return fmt.Errorf("parse args: %w", err)
	}

	if err := dev.StartCapture(usePacketFolding.Value()); err != nil {
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
