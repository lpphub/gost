package monitor

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/felixge/fgprof"
)

func StartPprof() {
	http.DefaultServeMux.Handle("/debug/fgprof", fgprof.Handler())

	go func() {
		if err := http.ListenAndServe(":6060", nil); err != nil {
			panic("go profiler server start error: " + err.Error())
		}
	}()
}
