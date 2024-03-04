package function

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const Version = "0.1.1"

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
	var it int64 = 0
	if params.Has("it") {
		it, err = strconv.ParseInt(params.Get("it"), 10, 64)
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
	rit := int64(0)
	for ; rit < it || time.Now().Sub(tb0).Nanoseconds() < tb; rit++ {
	}
	rtb := time.Now().Sub(tb0)
	rts := tb0.Sub(ts0)
	rtf := time.Now()
	rdt := rtf.Sub(rt0)

	res := map[string]string{}
	res["rt0"] = strconv.FormatInt(rt0.UnixNano(), 10)
	res["rtb"] = strconv.FormatInt(rtb.Nanoseconds(), 10)
	res["rit"] = strconv.FormatInt(rit, 10)
	res["rts"] = strconv.FormatInt(rts.Nanoseconds(), 10)
	res["rdt"] = strconv.FormatInt(rdt.Nanoseconds(), 10)
	res["rtf"] = strconv.FormatInt(rtf.UnixNano(), 10)

	r, err := json.Marshal(res)
	if err != nil {
		http.Error(resp, err.Error(), 500)
		return
	}
	resp.Header().Add("Content-Type", "plain/text")
	resp.Header().Add("X-Request-ID", params.Get("id"))
	resp.Header().Add("Version", Version)
	resp.WriteHeader(200)
	_, err = fmt.Fprintf(resp, string(r))
	if err != nil {
		http.Error(resp, err.Error(), 500)
		return
	}
}
