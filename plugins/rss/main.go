package main

import (
	"encoding/json"

	"github.com/gavmor/wasm-microkernel/guest"
)

func init() {
	guest.Register(func(reqJSON string) (string, error) {
		var req struct {
			XML string `json:"xml"`
		}
		if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
			return "", err
		}
		p, eps, err := parseRSS(req.XML)
		if err != nil {
			return "", err
		}
		out, _ := json.Marshal(map[string]any{"podcast": p, "episodes": eps})
		return string(out), nil
	})
}

func main() {}
