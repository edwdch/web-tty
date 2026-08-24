package session

import (
	"encoding/json"
	"fmt"
)

const (
	CmdInput  byte = '0'
	CmdResize byte = '1'
	CmdPause  byte = '2'
	CmdResume byte = '3'
	CmdPing   byte = '4'
	CmdHello  byte = '{'

	MsgOutput byte = '0'
	MsgTitle  byte = '1'
	MsgInfo   byte = '2'

	MinSize = 1
	MaxSize = 1000
)

type Size struct {
	Columns uint16 `json:"columns"`
	Rows    uint16 `json:"rows"`
}

func (s Size) Validate() error {
	if s.Columns < MinSize || s.Columns > MaxSize || s.Rows < MinSize || s.Rows > MaxSize {
		return fmt.Errorf("invalid size %dx%d", s.Columns, s.Rows)
	}
	return nil
}

func ParseHello(msg []byte) (Size, error) {
	if len(msg) == 0 || msg[0] != CmdHello {
		return Size{}, fmt.Errorf("invalid hello")
	}
	var sz Size
	if err := json.Unmarshal(msg, &sz); err != nil {
		return Size{}, err
	}
	if err := sz.Validate(); err != nil {
		return Size{}, err
	}
	return sz, nil
}

func ParseResize(msg []byte) (Size, error) {
	if len(msg) < 2 || msg[0] != CmdResize {
		return Size{}, fmt.Errorf("invalid resize")
	}
	var sz Size
	if err := json.Unmarshal(msg[1:], &sz); err != nil {
		return Size{}, err
	}
	if err := sz.Validate(); err != nil {
		return Size{}, err
	}
	return sz, nil
}

func OutputFrame(data []byte) []byte {
	out := make([]byte, 1+len(data))
	out[0] = MsgOutput
	copy(out[1:], data)
	return out
}

func InfoFrame(id string) []byte {
	body, err := json.Marshal(struct {
		ID string `json:"id"`
	}{ID: id})
	if err != nil {
		body = []byte(`{"id":""}`)
	}
	out := make([]byte, 1+len(body))
	out[0] = MsgInfo
	copy(out[1:], body)
	return out
}

func TitleFrame(title string) []byte {
	body := []byte(title)
	out := make([]byte, 1+len(body))
	out[0] = MsgTitle
	copy(out[1:], body)
	return out
}
