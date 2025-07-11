package ag_conf

import (
	"fmt"
	"strings"
	"testing"
)

func TestFindPlaceholderEndIndex(t *testing.T) {
	tests := []struct {
		name     string
		buf      string
		prefix   string
		suffix   string
		startIdx int
		want     int
	}{
		{
			name:     "simple placeholder",
			buf:      "prefix${key}suffix",
			prefix:   "${",
			suffix:   "}",
			startIdx: 6,  // points to ${
			want:     11, // points to }
		},
		{
			name:     "nested placeholder",
			buf:      "prefix${outer${inner}}suffix",
			prefix:   "${",
			suffix:   "}",
			startIdx: 6,
			want:     21,
		},
		{
			name:     "no matching suffix",
			buf:      "prefix${key",
			prefix:   "${",
			suffix:   "}",
			startIdx: 6,
			want:     -1,
		},
		{
			name:     "empty between prefix and suffix",
			buf:      "prefix${}suffix",
			prefix:   "${",
			suffix:   "}",
			startIdx: 6,
			want:     8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			si := strings.Index(tt.buf, tt.prefix)
			fmt.Printf("name: %s\n", tt.name)
			helper := NewPropertyPlaceholderHelper(tt.prefix, tt.suffix, ":", false)
			got := helper.findPlaceholderEndIndex(tt.buf, tt.startIdx)
			fmt.Printf("tt.startIdx: %d, si: %d\n", tt.startIdx, si)
			fmt.Printf("tt.want: %d, got: %d\n", tt.want, got)
			if got != tt.want {
				t.Errorf("findPlaceholderEndIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}
