package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"
)

func listen(conn net.Conn, onSound func(string)) error {
	if _, err := io.WriteString(conn, `{"id":1,"method":"objects/subscribe","params":{"objects":{"print_stats":["state"]},"response_template":{}}}`+"\x03"); err != nil {
		return err
	}
	type update struct {
		Status struct {
			PrintStats struct {
				State string `json:"state"`
			} `json:"print_stats"`
		} `json:"status"`
	}
	reader := bufio.NewReader(conn)
	previous := ""
	for {
		frame, err := reader.ReadBytes(3)
		if err != nil {
			return err
		}
		var message struct {
			ID     int             `json:"id"`
			Params update          `json:"params"`
			Result update          `json:"result"`
			Error  json.RawMessage `json:"error"`
		}
		if err := json.Unmarshal(frame[:len(frame)-1], &message); err != nil {
			return err
		}
		if len(message.Error) > 0 && string(message.Error) != "null" {
			return fmt.Errorf("UDS: %s", message.Error)
		}
		state := message.Params.Status.PrintStats.State
		if message.ID == 1 {
			// An initial push can precede the subscription reply; do not rewind it.
			if previous != "" {
				continue
			}
			state = message.Result.Status.PrintStats.State
		} else if message.ID != 0 {
			continue
		}
		if state == "" || state == previous {
			continue
		}
		fmt.Printf("%s %s -> %s\n", time.Now().Format("15:04:05"), previous, state)
		// First state is a baseline; "resumed" is a sound, not a UDS state.
		if previous != "" {
			if state == "printing" && previous == "paused" {
				onSound("resumed")
			} else if state == "printing" || state == "paused" || state == "complete" || state == "cancelled" {
				onSound(state)
			}
		}
		previous = state
	}
}
