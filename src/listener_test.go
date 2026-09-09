package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"net"
	"reflect"
	"testing"
)

func TestListen(t *testing.T) {
	for _, tc := range []struct {
		name   string
		frames []string
		want   []string
	}{
		{"lifecycle", []string{
			`{"id":1,"result":{"status":{"print_stats":{"state":"cancelled"}}}}`,
			`{"id":0,"params":{"status":{"print_stats":{"state":"printing"}}}}`,
			`{"id":0,"params":{"status":{"print_stats":{"state":"printing"}}}}`,
			`{"id":0,"params":{"status":{"print_stats":{"machine_sub_status":2501}}}}`,
			`{"id":0,"params":{"status":{"print_stats":{"state":"paused"}}}}`,
			`{"id":0,"params":{"status":{"pause_resume":{"is_paused":true}}}}`,
			`{"id":0,"params":{"status":{"print_stats":{"state":"printing"}}}}`,
			`{"id":0,"report":{"message":"Temperature: 200\u00b0C"}}`,
			`{"id":0,"params":{"status":{"print_stats":{"state":"cancelled"}}}}`,
			`{"id":0,"params":{"status":{"print_stats":{"state":"cancelled"}}}}`,
			`{"id":0,"params":{"status":{"print_stats":{"state":"printing"}}}}`,
			`{"id":0,"params":{"status":{"print_stats":{"state":"complete"}}}}`,
		}, []string{"printing", "paused", "resumed", "cancelled", "printing", "complete"}},
		{"push_before_reply", []string{
			`{"id":0,"params":{"status":{"print_stats":{"state":"printing"}}}}`,
			`{"id":1,"result":{"status":{"print_stats":{"state":"standby"}}}}`,
			`{"id":0,"params":{"status":{"print_stats":{"state":"printing"}}}}`,
		}, nil},
		{"initial_paused", []string{
			`{"id":1,"result":{"status":{"print_stats":{"state":"paused"}}}}`,
			`{"id":0,"params":{"status":{"print_stats":{"state":"printing"}}}}`,
			`{"id":0,"params":{"status":{"print_stats":{"state":"printing"}}}}`,
			`{"id":0,"params":{"status":{"print_stats":{"state":"error"}}}}`,
		}, []string{"resumed"}},
		{"stop_while_paused", []string{
			`{"id":1,"result":{"status":{"print_stats":{"state":"paused"}}}}`,
			`{"id":0,"params":{"status":{"print_stats":{"state":"cancelled"}}}}`,
		}, []string{"cancelled"}},
		{"initial_cancelled", []string{
			`{"id":1,"result":{"status":{"print_stats":{"state":"cancelled"}}}}`,
			`{"id":0,"params":{"status":{"print_stats":{"state":"cancelled"}}}}`,
		}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, server := net.Pipe()
			defer client.Close()
			defer server.Close()
			go func() {
				defer server.Close()
				request, err := bufio.NewReader(server).ReadBytes(3)
				if err != nil || !json.Valid(request[:len(request)-1]) {
					t.Error("invalid subscription request", err)
					return
				}
				for _, frame := range tc.frames {
					// Split every frame to exercise stream framing, including UTF-8.
					for _, b := range []byte(frame + "\x03") {
						if _, err := server.Write([]byte{b}); err != nil {
							return
						}
					}
				}
			}()
			var got []string
			err := listen(client, func(state string) { got = append(got, state) })
			if !errors.Is(err, io.EOF) {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("sounds = %v; want %v", got, tc.want)
			}
		})
	}
}
