package agent

import (
	"slices"
	"testing"
)

func TestBuildFFmpegArgsOmitsReFlag(t *testing.T) {
	args := buildFFmpegArgs("rtsp://cam/stream", "rtp://127.0.0.1:10000", "rtp://127.0.0.1:10001")
	if slices.Contains(args, "-re") {
		t.Fatal("buildFFmpegArgs must not include -re for live RTSP sources")
	}
}

func TestBuildFFmpegArgsStructure(t *testing.T) {
	args := buildFFmpegArgs("rtsp://cam/stream", "rtp://127.0.0.1:10000", "rtp://127.0.0.1:10001")

	// Verify key arguments are present in the correct order.
	want := []string{
		"-hide_banner",
		"-loglevel", "error",
		"-rtsp_transport", "tcp",
		"-i", "rtsp://cam/stream",
		"-c:v", "copy",
		"-an",
		"-f", "rtp",
		"rtp://127.0.0.1:10000",
		"-c:a", "libopus",
		"-ar", "48000",
		"-ac", "2",
		"-vn",
		"-f", "rtp",
		"rtp://127.0.0.1:10001",
	}

	if len(args) != len(want) {
		t.Fatalf("arg count mismatch: got %d, want %d", len(args), len(want))
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("arg[%d]: got %q, want %q", i, args[i], want[i])
		}
	}
}

func TestRewritePayloadType(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		rewrite uint8
		want    []byte
	}{
		{
			name:    "rewrite pt 97 to 111",
			input:   []byte{0x80, 0x61, 0x00, 0x00}, // PT=97 (0x61 & 0x7F)
			rewrite: 111,
			want:    []byte{0x80, 0x6F, 0x00, 0x00}, // PT=111 (0x6F & 0x7F)
		},
		{
			name:    "preserve marker bit",
			input:   []byte{0x80, 0xE1, 0x00, 0x00}, // M=1, PT=97
			rewrite: 111,
			want:    []byte{0x80, 0xEF, 0x00, 0x00}, // M=1, PT=111
		},
		{
			name:    "short packet no rewrite",
			input:   []byte{0x80}, // only 1 byte
			rewrite: 111,
			want:    []byte{0x80}, // unchanged
		},
		{
			name:    "rewrite to pt 0",
			input:   []byte{0x80, 0x61, 0x00, 0x00},
			rewrite: 0,
			want:    []byte{0x80, 0x00, 0x00, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkt := make([]byte, len(tt.input))
			copy(pkt, tt.input)
			rewritePayloadType(pkt, tt.rewrite)
			if len(pkt) != len(tt.want) {
				t.Fatalf("length mismatch: got %d, want %d", len(pkt), len(tt.want))
			}
			for i := range tt.want {
				if pkt[i] != tt.want[i] {
					t.Fatalf("byte %d: got 0x%02X, want 0x%02X", i, pkt[i], tt.want[i])
				}
			}
		})
	}
}
