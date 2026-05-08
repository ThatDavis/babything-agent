package main

import "testing"

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
