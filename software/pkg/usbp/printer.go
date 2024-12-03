package usbp

import (
	"fmt"
	"io"
)

func Print(in []byte, w io.Writer) {
	n := len(in)
	rowcount := 0
	stop := (n / 8) * 8
	k := 0
	for i := 0; i <= stop; i += 8 {
		k++
		switch {
		case i+8 < n:
			rowcount = 8
		case k*8 < n:
			rowcount = 0
		default:
			rowcount = n % 8
		}

		for j := 0; j < rowcount; j++ {
			_, _ = fmt.Fprintf(w, "%02x  ", in[i+j])
		}
		for j := rowcount; j < 8; j++ {
			_, _ = fmt.Fprint(w, "    ")
		}

		_, _ = fmt.Fprintf(w, "  '%s'\n", viewString(in[i:(i+rowcount)]))
	}

}

func viewString(b []byte) string {
	r := []rune(string(b))
	for i := range r {
		if r[i] < 32 || r[i] > 126 {
			r[i] = '.'
		}
	}
	return string(r)
}
