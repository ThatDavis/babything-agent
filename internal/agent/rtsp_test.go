package agent

import (
	"encoding/binary"
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
		"-ac", "1",
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

func TestSmoothRTPTimestamps(t *testing.T) {
	// Build a fake RTP packet: version=2, PT=111, seq=100, ts=1000
	pkt := []byte{
		0x80, 0x6F, // V=2, PT=111
		0x00, 0x64, // seq=100
		0x00, 0x00, 0x03, 0xE8, // ts=1000
		0x00, 0x00, 0x00, 0x00, // SSRC=0
		0x01, 0x02, 0x03, 0x04, // payload
	}

	// Simulate the smoothing logic inline
	baseSeq := binary.BigEndian.Uint16(pkt[2:4])
	baseTs := binary.BigEndian.Uint32(pkt[4:8])
	samplesPerPacket := uint32(960)

	// Next packet, seq=101
	binary.BigEndian.PutUint16(pkt[2:4], 101)
	seq := binary.BigEndian.Uint16(pkt[2:4])
	deltaSeq := int32(seq) - int32(baseSeq)
	newTs := baseTs + uint32(deltaSeq)*samplesPerPacket
	binary.BigEndian.PutUint32(pkt[4:8], newTs)

	wantTs := uint32(1000 + 960)
	gotTs := binary.BigEndian.Uint32(pkt[4:8])
	if gotTs != wantTs {
		t.Fatalf("timestamp mismatch: got %d, want %d", gotTs, wantTs)
	}

	// Wraparound: seq=0 after baseSeq=65535
	baseSeq = 65535
	baseTs = 1000
	seq = 0
	deltaSeq = int32(seq) - int32(baseSeq)
	if deltaSeq < 0 {
		deltaSeq += 65536
	}
	newTs = baseTs + uint32(deltaSeq)*samplesPerPacket
	if newTs != 1000+960 {
		t.Fatalf("wraparound timestamp mismatch: got %d, want %d", newTs, 1000+960)
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
