package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/textproto"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const S_NS = 1000000000

const Version = "0-1-1"
const ClientId = "tracexec-" + Version

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var _TraceFilename = getEnv("TRACE", "invokes.csv")
var DutyCycle, _e1 = strconv.ParseFloat(getEnv("DUTY", "0.25"), 64)
var BaseURL = getEnv("URL", "http://200.144.244.220:10080/")
var BeginAt, _e2 = time.Parse(time.RFC3339, getEnv("BEGIN",
	time.Now().Truncate(time.Minute).Add(2*time.Minute).Format(time.RFC3339)))
var DbgFunc = getEnv("DBGFUNC", "")
var OutDir = getEnv("OUTDIR", "logs")
var MaxTime = time.Unix(1<<63-62135596801, 999999999)
var InitialRecordID, _e3 = strconv.ParseInt(getEnv("INITRID", "0"), 10, 64)
var FinalRecordID, _e4 = strconv.ParseInt(getEnv("ENDRID", "-1"), 10, 64)

var _baseLogFilename = BeginAt.Format(time.RFC3339)
var _OutFilename = OutDir + "/" + _baseLogFilename + ".out"
var _ErrFilename = OutDir + "/" + _baseLogFilename + ".err"
var _DbgFilename = OutDir + "/" + _baseLogFilename + ".dbg"
var _OutFile *os.File = nil
var _ErrFile *os.File = nil
var _DbgFile *os.File = nil
var _err error = nil

func TimestampNowAsString() string {
	return time.Now().Format(time.RFC3339)
}

func saveSuccess(wg *sync.WaitGroup, message string) {
	defer wg.Done()
	_, _err = _OutFile.WriteString(TimestampNowAsString() + "\t" + message + "\n")
	if _err != nil {
		saveError(_err, false, nil)
	}
}

func saveError(errArg error, exit bool, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	var message map[string]any
	message["error"] = true
	message["arg"] = errArg.Error()
	message["exit"] = exit
	msgJson, err := json.Marshal(message)
	if err != nil {
		if exit {
			msgJson = []byte("{\"error\":True,\"exit\":True,\"arg\":\"Error encoding JSON for another error report\"}")
		} else {
			msgJson = []byte("{\"error\":True,\"exit\":False,\"arg\":\"Error encoding JSON for another error report\"}")
		}
	}

	if _ErrFile != nil {
		_, _err = _ErrFile.WriteString(TimestampNowAsString() + "\t" + string(msgJson) + "\n")
		if _err != nil {
			fmt.Println(TimestampNowAsString(), message)
		}
	} else {
		fmt.Println(TimestampNowAsString(), message)
	}
	if exit {
		fmt.Println(TimestampNowAsString(), fmt.Errorf("[FATAL]:  \"process exiting with non-zero status\""))
		os.Exit(1)
	}
}

func saveDebug(message string) {
	if _DbgFile != nil {
		_, _err = _DbgFile.WriteString(TimestampNowAsString() + "\t" + message + "\n")
		if _err == nil {
			return
		}
	}
	fmt.Println(TimestampNowAsString(), "[DEBUG]:  "+message)
}

func traceGen(rid string) *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		ConnectStart: func(network, addr string) {
			a := map[string]string{"event": "ConnectStart", "rid": rid, "network": network, "addr": addr}
			s, err := json.Marshal(a)
			if err != nil {
				go saveError(err, false, nil)
				return
			}
			go saveDebug(string(s))
		},
		DNSStart: func(info httptrace.DNSStartInfo) {
			a := map[string]string{"event": "DNSStart", "rid": rid, "host": info.Host}
			s, err := json.Marshal(a)
			if err != nil {
				go saveError(err, false, nil)
				return
			}
			go saveDebug(string(s))
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			a := map[string]any{"event": "DNSDone", "rid": rid, "coalesced": info.Coalesced}
			if info.Err != nil {
				a["error"] = info.Err.Error()
			}
			var addrs []map[string]string
			for _, v := range info.Addrs {
				b := map[string]string{}
				b["IP"] = v.IP.String()
				b["Zone"] = v.Zone
				addrs = append(addrs, b)
			}
			a["addrs"] = addrs
			jsonA, err := json.Marshal(a)
			if err != nil {
				go saveError(err, false, nil)
				return
			}
			go saveDebug(string(jsonA))
		},
		GetConn: func(hostPort string) {
			a := map[string]string{"event": "GetConn", "rid": rid, "hostPort": hostPort}
			jsonA, err := json.Marshal(a)
			if err != nil {
				go saveError(err, false, nil)
				return
			}
			go saveDebug(string(jsonA))
		},
		GotConn: func(info httptrace.GotConnInfo) {
			a := map[string]any{}
			a["event"] = "GotConn"
			a["reused"] = info.Reused
			a["wasIdle"] = info.WasIdle
			a["local"] = info.Conn.LocalAddr().String()
			a["remote"] = info.Conn.RemoteAddr().String()
			jsonA, err := json.Marshal(a)
			if err != nil {
				go saveError(err, false, nil)
				return
			}
			go saveDebug(string(jsonA))

		},
		WroteHeaderField: func(key string, value []string) {
			a := map[string]any{"event": "WroteHeaderField", "rid": rid, "value": map[string][]string{key: value}}
			jsonA, err := json.Marshal(a)
			if err != nil {
				go saveError(err, false, nil)
				return
			}
			go saveDebug(string(jsonA))
		},
		WroteHeaders: func() {
			a := map[string]string{"event": "WroteHeaders", "rid": rid}
			jsonA, _ := json.Marshal(a)
			go saveDebug(string(jsonA))
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			a := map[string]string{"event": "WroteRequest", "rid": rid}
			if info.Err != nil {
				a["error"] = info.Err.Error()
			}
			jsonA, err := json.Marshal(a)
			if err != nil {
				go saveError(err, false, nil)
				return
			}
			go saveDebug(string(jsonA))

		},
		GotFirstResponseByte: func() {
			a := map[string]string{"event": "GotFirstResponseByte", "rid": rid}
			jsonA, _ := json.Marshal(a)
			go saveDebug(string(jsonA))
		},
		Wait100Continue: func() {
			a := map[string]string{"event": "Wait100Continue", "rid": rid}
			jsonA, _ := json.Marshal(a)
			go saveDebug(string(jsonA))
		},
		Got100Continue: func() {
			a := map[string]string{"event": "Got100Continue", "rid": rid}
			jsonA, _ := json.Marshal(a)
			go saveDebug(string(jsonA))
		},
		Got1xxResponse: func(code int, header textproto.MIMEHeader) error {
			a := map[string]any{"event": "Got1xxResponse", "code": code, "headers": header}
			jsonA, err := json.Marshal(a)
			if err != nil {
				go saveError(err, false, nil)
				return nil
			}
			go saveDebug(string(jsonA))
			return nil
		},
		PutIdleConn: func(err error) {
			a := map[string]any{"event": "PutIdleConn", "rid": rid}
			if err != nil {
				a["error"] = err.Error()
			}
			jsonA, err := json.Marshal(a)
			if err != nil {
				go saveError(err, false, nil)
				return
			}
			go saveDebug(string(jsonA))
		},
		ConnectDone: func(network, addr string, err error) {
			a := map[string]any{"event": "ConnectDone", "rid": rid, "addr": addr}
			if err != nil {
				a["error"] = err.Error()
			}
			jsonA, err := json.Marshal(a)
			if err != nil {
				go saveError(err, false, nil)
				return
			}
			go saveDebug(string(jsonA))
		},
	}
}

func main() {

	fmt.Println("Tracexec Simulator")
	fmt.Println("---")
	fmt.Println("VERSION: ", Version)
	fmt.Println("URL: ", BaseURL)
	fmt.Println("TRACE: ", _TraceFilename)
	fmt.Println("RANGE: ", "["+strconv.FormatInt(InitialRecordID, 10)+":"+strconv.FormatInt(FinalRecordID, 10)+"]")
	fmt.Println("DUTY: ", DutyCycle)
	fmt.Println("BEGIN: ", BeginAt.Format(time.RFC3339))
	fmt.Println("OUT: ", _OutFilename)
	fmt.Println("ERR: ", _ErrFilename)
	fmt.Println("DBG: ", _DbgFilename)
	if len(DbgFunc) > 0 {
		fmt.Println("DBGFUNC: " + DbgFunc)
	}
	fmt.Println("---")

	_ErrFile, _err = os.OpenFile(_ErrFilename, os.O_CREATE|os.O_RDWR, 0664)
	if _err != nil {
		log.Fatal(_err)
	}
	if _e1 != nil || _e2 != nil || _e3 != nil || _e4 != nil {
		saveError(_err, true, nil)
	}
	if BeginAt.Unix() < time.Now().Unix() {
		saveError(fmt.Errorf("begin at past time not allowed"), true, nil)
	}
	_OutFile, _err = os.OpenFile(_OutFilename, os.O_CREATE|os.O_RDWR, 0664)
	if _err != nil {
		saveError(_err, true, nil)
	}
	_DbgFile, _err = os.OpenFile(_DbgFilename, os.O_CREATE|os.O_RDWR, 0664)
	if _err != nil {
		saveError(_err, true, nil)
	}

	_TraceFile, _err := os.ReadFile(_TraceFilename)
	if _err != nil {
		saveError(_err, true, nil)
	}

	wg := sync.WaitGroup{}
	tr := &http.Transport{
		MaxIdleConns:          0,
		MaxIdleConnsPerHost:   math.MaxInt64,
		MaxConnsPerHost:       0,
		IdleConnTimeout:       0,
		ResponseHeaderTimeout: 18 * time.Minute,
		ExpectContinueTimeout: 18 * time.Minute,
		DisableKeepAlives:     false,
		DialContext: (&net.Dialer{
			Timeout:   18 * time.Minute,
			KeepAlive: 15 * time.Second,
			Deadline:  MaxTime,
		}).DialContext,
	}
	client := &http.Client{
		Timeout:   0,
		Transport: tr,
	}
	ctx := context.Background()

	requests := map[int]*http.Request{}
	data := map[int]map[string]string{}
	lines := strings.Split(string(_TraceFile), "\n")
	if FinalRecordID == -1 {
		FinalRecordID = int64(len(lines) - 2)
	}
	for i, line := range lines[InitialRecordID+1 : FinalRecordID+2] {
		row := strings.Split(line, ",")
		if len(row) != 5 {
			continue
		}
		rid := row[0]
		t0, _ := strconv.ParseFloat(row[1], 64)
		fid := row[2]
		dur, _ := strconv.ParseFloat(row[3], 64)
		duration := dur * S_NS
		tb := uint64(DutyCycle * duration)
		ts := uint64(duration) - tb
		req, err := http.NewRequestWithContext(ctx, "GET", BaseURL, nil)
		if err != nil {
			saveError(err, true, nil)
		}
		req = req.WithContext(httptrace.WithClientTrace(req.Context(), traceGen(rid)))
		req.Header.Add("Content-Type", "text/plain")
		req.Header.Add("Accept", "*/*")
		if len(DbgFunc) == 0 {
			req.Host = "simtask-" + fid + ".default.knative.dev"
		} else {
			req.Host = DbgFunc + ".default.knative.dev"
		}

		query := map[string]string{}
		data[i] = query
		query["cl"] = ClientId
		query["fid"] = fid
		query["id"] = rid
		query["ts"] = strconv.FormatUint(ts, 10)
		query["tb"] = strconv.FormatUint(tb, 10)
		query["t0"] = strconv.FormatUint(uint64(t0*S_NS), 10)
		q := req.URL.Query()
		for k, v := range query {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
		requests[i] = req
	}
	fmt.Println("Requests pre generated. Start scheduled for " +
		BeginAt.Format(time.RFC3339))
	fmt.Println("---")
	wg.Add(len(requests))
	for i := 0; i < len(requests); i++ {
		row := data[i]
		t0, _ := strconv.ParseUint(row["t0"], 10, 64)
		var dt int64
		for dt = time.Now().Sub(BeginAt).Nanoseconds() - int64(t0); dt < 0; dt = time.Now().Sub(BeginAt).Nanoseconds() - int64(t0) {
		}
		go func(rid int) {
			req := requests[rid]
			resp, err := client.Do(req)
			if err != nil || resp.StatusCode != 200 {
				if err == nil {
					err = fmt.Errorf("{\"rid\":%d,\"code\":%d,\"message\":\"%s\"}", rid, resp.StatusCode, resp.Status)
				}
				go saveError(err, false, &wg)
				return
			}
			go func(query map[string]string, resp *http.Response) {
				ret, err := io.ReadAll(resp.Body)
				if err != nil {
					go saveError(err, false, &wg)
					return
				}
				_ = resp.Body.Close()

				retMap := map[string]string{}
				err = json.Unmarshal(ret, &retMap)
				if err != nil {
					go saveError(err, false, &wg)
					return
				}
				for k, v := range query {
					retMap[k] = v
				}
				retMap["Tf"] = strconv.FormatInt(time.Now().UnixNano(), 10)
				retMapBuf, err := json.Marshal(retMap)
				if err != nil {
					go saveError(err, false, &wg)
					return
				}
				go saveSuccess(&wg, string(retMapBuf))
			}(row, resp)
		}(i)
	}
	wg.Wait()
}
