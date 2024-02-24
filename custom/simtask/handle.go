package function

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

func Handle(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
	te0 := time.Now()
	params := req.URL.Query()
	if !params.Has("cl") {
		resp.WriteHeader(200)
		_, _ = resp.Write([]byte(""))
		return
	}

	rt0 := time.Now().UnixNano()
	t0, err := strconv.ParseUint(params.Get("t0"), 10, 64)
	if err != nil {
		http.Error(resp, "bad 't0' parameter", 400)
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
	rte := time.Now().Sub(te0)

	res := map[string]any{}
	res["cl"] = params.Get("cl")
	res["rid"] = params.Get("id")
	res["fid"] = params.Get("fid")
	res["t0"] = t0
	res["rt0"] = rt0
	res["rtb"] = rtb
	res["tb"] = tb
	res["rts"] = rts
	res["ts"] = ts
	res["rte"] = rte

	r, err := json.Marshal(res)
	if err != nil {
		resp.WriteHeader(500)
		_, err := resp.Write([]byte(err.Error()))
		if err != nil {
			http.Error(resp, err.Error(), 500)
			return
		}
	}
	resp.WriteHeader(200)
	_, err = resp.Write(r)
	if err != nil {
		http.Error(resp, err.Error(), 500)
		return
	}
}
