package commands

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
	"github.com/gopacket/gopacket/pcapgo"
	"github.com/spf13/cobra"

	"github.com/buglloc/mr.duppl/software/pkg/dupplcap"
	"github.com/buglloc/mr.duppl/software/pkg/usbp"
)

var dumpArgs struct {
	Interface string
	NoFolding bool
	Output    string
}

var dumpCmd = &cobra.Command{
	Use:           "dump",
	SilenceUsage:  true,
	SilenceErrors: true,
	Short:         "Dump Mr.Duppl capture",
	RunE: func(_ *cobra.Command, _ []string) error {
		dev, err := dupplcap.NewDevice(dumpArgs.Interface)
		if err != nil {
			return fmt.Errorf("opening device: %w", err)
		}
		defer func() { _ = dev.Close() }()

		if err := dev.StartCapture(!dumpArgs.NoFolding); err != nil {
			return fmt.Errorf("start capture: %w", err)
		}
		defer func() { _ = dev.StopCapture() }()

		if len(dumpArgs.Output) != 0 {
			out, err := os.Create(dumpArgs.Output)
			if err != nil {
				return fmt.Errorf("creating output file: %w", err)
			}
			defer func() { _ = out.Close() }()

			return dumpCapture(dev, out)
		}

		return printCapture(dev)
	},
}

func dumpCapture(dev *dupplcap.Device, out io.Writer) error {
	w, err := pcapgo.NewNgWriterInterface(
		out,
		pcapgo.NgInterface{
			Name:       dev.Iface(),
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
	defer func() { _ = w.Flush() }()

	for {
		packet, err := dev.Packet()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

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

		usbp.Print(packet, os.Stdout)
	}
}

func printCapture(dev *dupplcap.Device) error {
	for {
		packet, err := dev.Packet()
		if err != nil {
			return fmt.Errorf("read packet: %w", err)
		}

		usbp.Print(packet, os.Stdout)
	}
}

func init() {
	flags := dumpCmd.PersistentFlags()
	flags.StringVarP(&dumpArgs.Interface, "iface", "i", "", "interface to use")
	flags.StringVarP(&dumpArgs.Output, "out", "o", "", "write PcapNG file")
	flags.BoolVar(&dumpArgs.NoFolding, "no-packet-folding", false, "disable packet folding")
}
