package slp

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func Test_encodePingOld(t *testing.T) {
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
			frame: "fe",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			gotErr := encodePingOld(buf, tt.args.host, tt.args.port)

			if (tt.err != nil) != (gotErr != nil) {
				t.Errorf("expected %v, got %v", tt.err, gotErr)
			}
			formatted := fmt.Sprintf("%x", buf.Bytes())
			assert.Equal(t, tt.frame, formatted)
		})
	}
}
