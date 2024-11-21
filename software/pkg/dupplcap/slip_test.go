package dupplcap

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

var readData = []struct {
	data     []byte
	expected []byte
	isPrefix bool
	err      error
}{
	{[]byte{}, []byte{}, true, io.EOF},
	// All the slipEnd are received till EOF or data
	{[]byte{slipEnd, slipEnd, slipEnd, slipEnd}, []byte{}, true, io.EOF},
	{[]byte{slipEnd, slipEnd, 1, slipEnd}, []byte{1}, false, nil},
	// Properly terminated data
	{[]byte{1, 2, 3, slipEnd}, []byte{1, 2, 3}, false, nil},
	{[]byte{1, 2, 3, slipEnd, 4, 5, 6}, []byte{1, 2, 3}, false, nil},
	{[]byte{slipEsc, slipEscEsc, slipEnd}, []byte{slipEsc}, false, nil},
	{[]byte{slipEsc, slipEscEnd, slipEnd}, []byte{slipEnd}, false, nil},
	// Non terminated data
	{[]byte{1, 2, 3}, []byte{1, 2, 3}, true, io.EOF},
	{[]byte{slipEsc, slipEscEsc}, []byte{slipEsc}, true, io.EOF},
	{[]byte{slipEsc, slipEscEnd}, []byte{slipEnd}, true, io.EOF},
	// Bad control sequences
	{[]byte{1, slipEscEsc, 3}, []byte{1, slipEscEsc, 3}, true, io.EOF},
	{[]byte{1, slipEscEnd, 3}, []byte{1, slipEscEnd, 3}, true, io.EOF},
	{[]byte{1, slipEsc, 3}, []byte{1, 3}, true, io.EOF},
}

var writeData = []struct {
	data     []byte
	expected []byte
	err      error
}{
	// Just data. Starts with slipEnd and ends with slipEnd
	{[]byte{1, 2, 3}, []byte{1, 2, 3, slipEnd}, nil},
	// Escape sequences
	{[]byte{slipEnd}, []byte{slipEsc, slipEscEnd, slipEnd}, nil},
	{[]byte{slipEsc}, []byte{slipEsc, slipEscEsc, slipEnd}, nil},
	{[]byte{slipEscEnd}, []byte{slipEscEnd, slipEnd}, nil},
	{[]byte{slipEscEsc}, []byte{slipEscEsc, slipEnd}, nil},
	{[]byte{slipEnd, slipEsc}, []byte{slipEsc, slipEscEnd, slipEsc, slipEscEsc, slipEnd}, nil},
}

func TestRead(t *testing.T) {
	for i, d := range readData {
		t.Run(fmt.Sprintf("read_%d", i), func(t *testing.T) {
			p, err := readSlipPacket(bytes.NewReader(d.data))
			if d.err != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, d.err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, d.expected, p)
		})
	}
}

func TestWrite(t *testing.T) {
	for i, d := range writeData {
		t.Run(fmt.Sprintf("write_%d", i), func(t *testing.T) {
			buf := &bytes.Buffer{}
			err := writeSlipPacket(buf, d.data)
			if d.err != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, d.err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, d.expected, buf.Bytes())
		})
	}
}

func TestWriteAndRead(t *testing.T) {
	for i, d := range writeData {
		t.Run(fmt.Sprintf("rw_%d", i), func(t *testing.T) {
			if d.err != nil {
				t.Skip()
				return
			}

			buf := &bytes.Buffer{}
			err := writeSlipPacket(buf, d.data)
			require.NoError(t, err)

			p, err := readSlipPacket(buf)
			require.NoError(t, err)
			require.Equal(t, d.data, p)
		})
	}
}
