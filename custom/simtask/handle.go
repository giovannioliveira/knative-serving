package function

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const Version = "0.1.0"

func Handle(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
	rt0 := time.Now()
	params := req.URL.Query()
	if !params.Has("cl") {
		resp.WriteHeader(200)
		_, _ = resp.Write([]byte(""))
		return
	}
	ts, err := strconv.ParseInt(params.Get("ts"), 10, 64)
	if err != nil {
		http.Error(resp, "bad 'ts' parameter", 400)
		return
	}
	var tb int64 = 0
	var it uint64 = 0
	if params.Has("it") {
		it, err = strconv.ParseUint(params.Get("it"), 10, 64)
		if err != nil {
			http.Error(resp, "bad 'it' parameter", 400)
			return
		}
	} else {
		tb, err = strconv.ParseInt(params.Get("tb"), 10, 64)
		if err != nil {
			http.Error(resp, "bad 'tb' parameter", 400)
			return
		}
	}
	ts0 := time.Now()
	if ts > 0 {
		time.Sleep(time.Duration(ts))
	}
	tb0 := time.Now()
	rit := uint64(0)
	for ; rit < it || time.Now().Sub(tb0).Nanoseconds() < tb; rit++ {
	}
	rtb := time.Now().Sub(tb0)
	rts := tb0.Sub(ts0)
	rtf := time.Now()
	rdt := rtf.Sub(rt0)

	res := map[string]any{}
	res["rt0"] = rt0.UnixNano()
	res["rtb"] = rtb.Nanoseconds()
	res["rts"] = rts.Nanoseconds()
	res["rdt"] = rdt.Nanoseconds()
	res["rtf"] = rtf.UnixNano()

	r, err := json.Marshal(res)
	if err != nil {
		resp.WriteHeader(500)
		_, err := resp.Write([]byte(err.Error()))
		if err != nil {
			http.Error(resp, err.Error(), 500)
			return
		}
	}
	resp.Header().Add("Content-Type", "application/json")
	resp.Header().Add("X-Request-ID", params.Get("id"))
	resp.Header().Add("X-Request-Function", strings.Split(req.Host, ".")[0])
	resp.Header().Add("Version", Version)
	resp.WriteHeader(200)
	_, err = fmt.Fprintf(resp, string(r))
	if err != nil {
		http.Error(resp, err.Error(), 500)
		return
	}
}
