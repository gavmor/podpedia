// Package kernel is the Podpedia-specific boot layer for the wasm-microkernel.
// It wires configuration from CLI flags into the microkernel. Everything below
// this — wazero, ABI, capability registration — lives in github.com/gavmor/wasm-microkernel.
package kernel

import (
	"context"
	"fmt"
	"io"

	"code.cloudfoundry.org/lager/v3"
	wmk "github.com/gavmor/wasm-microkernel/kernel"
)

// Kernel wraps the microkernel with Podpedia's URL config, which adapters
// thread through plugin requests so plugins own the API call logic.
type Kernel struct {
	*wmk.Kernel
	ollamaURL     string
	transcribeURL string
}

// New boots the microkernel with standard capabilities registered under
// the canonical WIT module name. The host app provides only config.
func New(ctx context.Context, logger lager.Logger, ollamaURL, transcribeURL string) (*Kernel, error) {
	k, err := wmk.New(ctx, wmk.Config{
		LogWriter: lagerWriter{logger},
	})
	if err != nil {
		return nil, fmt.Errorf("microkernel init: %w", err)
	}
	return &Kernel{Kernel: k, ollamaURL: ollamaURL, transcribeURL: transcribeURL}, nil
}

// lagerWriter bridges io.Writer to lager.Logger so plugin log messages
// appear in the structured application log.
type lagerWriter struct{ l lager.Logger }

func (w lagerWriter) Write(p []byte) (int, error) {
	w.l.Info("plugin", lager.Data{"msg": string(p)})
	return len(p), nil
}

// Ensure lagerWriter satisfies io.Writer at compile time.
var _ io.Writer = lagerWriter{}
