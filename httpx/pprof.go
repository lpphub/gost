package httpx

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/felixge/fgprof"
)

type pprofConfig struct {
	port int
}

type PprofOption func(*pprofConfig)

func WithPprofPort(port int) PprofOption {
	return func(c *pprofConfig) { c.port = port }
}

func StartPprof(opts ...PprofOption) {
	cfg := &pprofConfig{port: 6060}
	for _, o := range opts {
		o(cfg)
	}

	http.DefaultServeMux.Handle("/debug/fgprof", fgprof.Handler())
	go func() {
		addr := fmt.Sprintf(":%d", cfg.port)
		if err := http.ListenAndServe(addr, nil); err != nil {
			panic("go profiler server start error: " + err.Error())
		}
	}()
}
