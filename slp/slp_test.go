package slp

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func Test_encodePing(t *testing.T) {
	type args struct {
		host string
		port int
	}
	tests := []struct {
		name  string
		args  args
		err   error
		frame string
	}{
		{
			name: "from spec",
			args: args{
				host: "localhost",
				port: 25565,
			},
			err: nil,
			frame: "fe01fa000b004d0043007c0050006900" +
				"6e00670048006f0073007400194a0009" +
				"006c006f00630061006c0068006f0073" +
				"0074000063dd",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			gotErr := encodePing(buf, tt.args.host, tt.args.port)

			if (tt.err != nil) != (gotErr != nil) {
				t.Errorf("expected %v, got %v", tt.err, gotErr)
			}
			formatted := fmt.Sprintf("%x", buf.Bytes())
			assert.Equal(t, tt.frame, formatted)
		})
	}
}

func Test_readString(t *testing.T) {
	type args struct {
		conn io.Reader
	}
	tests := []struct {
		name     string
		args     args
		err      error
		expected string
	}{
		{
			name: "from spec",
			args: args{
				bytes.NewBuffer([]byte{0x00, 0xa7, 0x00, 0x31, 0x00, 0x00}),
			},
			err:      nil,
			expected: "§1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStr, gotErr := readString(tt.args.conn)
			if (tt.err != nil) != (gotErr != nil) {
				t.Errorf("expected %v, got %v", tt.err, gotErr)
			}
			assert.Equal(t, tt.expected, gotStr)
		})
	}
}
