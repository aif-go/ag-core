package toml

import (
	"reflect"
	"testing"
	"time"
)

func TestRead(t *testing.T) {
	// Parse expected time for nested tables test
	dobTime, _ := time.Parse(time.RFC3339, "1979-05-27T07:32:00Z")
	
	tests := []struct {
		name     string
		input    []byte
		want     map[string]interface{}
		wantErr  bool
	}{
		{
			name: "simple key-value",
			input: []byte(`title = "TOML Example"

[database]
server = "192.168.1.1"
ports = [8001, 8002, 8003]
connection_max = 5000
enabled = true`),
			want: map[string]interface{}{
				"title": "TOML Example",
				"database": map[string]interface{}{
					"server":         "192.168.1.1",
					"ports":          []interface{}{int64(8001), int64(8002), int64(8003)},
					"connection_max": int64(5000),
					"enabled":        true,
				},
			},
			wantErr: false,
		},
		{
			name: "nested tables",
			input: []byte(`[owner]
name = "Tom Preston-Werner"
dob = 1979-05-27T07:32:00Z

[database]
server = "192.168.1.1"
ports = [8001, 8002, 8003]
connection_max = 5000
enabled = true

[servers]

[servers.alpha]
ip = "10.0.0.1"
dc = "eqdc10"

[servers.beta]
ip = "10.0.0.2"
dc = "eqdc10"
role = "frontend"`),
			want: map[string]interface{}{
				"owner": map[string]interface{}{
					"name": "Tom Preston-Werner",
					"dob":  dobTime,
				},
				"database": map[string]interface{}{
					"server":         "192.168.1.1",
					"ports":          []interface{}{int64(8001), int64(8002), int64(8003)},
					"connection_max": int64(5000),
					"enabled":        true,
				},
				"servers": map[string]interface{}{
					"alpha": map[string]interface{}{
						"ip": "10.0.0.1",
						"dc": "eqdc10",
					},
					"beta": map[string]interface{}{
						"ip":   "10.0.0.2",
						"dc":   "eqdc10",
						"role": "frontend",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "empty input",
			input:   []byte(``),
			want:    map[string]interface{}{},
			wantErr: false,
		},
		{
			name:    "invalid TOML",
			input:   []byte(`title = "unclosed quote`),
			want:    nil,
			wantErr: true,
		},
		{
			name: "array of tables",
			input: []byte(`[[products]]
name = "Hammer"
sku = 738594937

[[products]]

[[products]]
name = "Nail"
sku = 284758393
color = "gray"`),
			want: map[string]interface{}{
				"products": []interface{}{
					map[string]interface{}{
						"name": "Hammer",
						"sku":  int64(738594937),
					},
					map[string]interface{}{},
					map[string]interface{}{
						"name":  "Nail",
						"sku":   int64(284758393),
						"color": "gray",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Read(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Read() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
