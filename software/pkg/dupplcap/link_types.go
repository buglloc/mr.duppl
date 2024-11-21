package dupplcap

import "fmt"

type LinkType int

const (
	LinkTypeUSBLowSpeed  LinkType = 293
	LinkTypeUSBFullSpeed LinkType = 294
	LinkTypeUSBHighSpeed LinkType = 295
)

func (l LinkType) Int() int {
	return int(l)
}

func (l LinkType) String() string {
	switch l {
	case LinkTypeUSBLowSpeed:
		return "USB_2_0_LOW_SPEED"
	case LinkTypeUSBFullSpeed:
		return "USB_2_0_FULL_SPEED"
	case LinkTypeUSBHighSpeed:
		return "USB_2_0_HIGH_SPEED"
	default:
		return fmt.Sprintf("TYPE_%d", l)
	}
}
