package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const Version = "0-1-0c"
const ClientId = "tracexec-" + Version

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var TraceFilename = getEnv("TRACE", "invokes.csv")
var DutyCycle, e1 = strconv.ParseFloat(getEnv("DUTY", ".25"), 64)
var IdleCycle = 1 - DutyCycle
var BaseURL = getEnv("URL", "http://10.4.0.143:10080/")
var BeginAt, e2 = time.Parse(time.RFC3339, getEnv("BEGIN",
	time.Now().Truncate(time.Minute).Add(2*time.Minute).Format(time.RFC3339)))
var DbgFunc = getEnv("DBGFUNC", "")
var OutDir = getEnv("OUTDIR", "logs")
var _baseLogFilename = BeginAt.Format(time.RFC3339)
var OutFilename = OutDir + "/" + _baseLogFilename + ".out"
var ErrFilename = OutDir + "/" + _baseLogFilename + ".err"
var OutFile *os.File = nil
var ErrFile *os.File = nil
var err error = nil

func saveSuccess(wg *sync.WaitGroup, message string) {
	defer wg.Done()
	_, err = OutFile.WriteString("[success]: " + message + "\n")
	if err != nil {
		saveError(err)
	}
}

func saveError(errArg error) {
	if ErrFile != nil {
		_, err = ErrFile.WriteString("[error]: " + errArg.Error() + "\n")
		if err != nil {
			fmt.Println("[error] :" + err.Error())
		}
	} else {
		fmt.Println("[error] :" + err.Error())
	}
}

func main() {
	fmt.Println("Tracexec Simulator")
	fmt.Println("---")
	fmt.Println("VERSION: ", Version)
	fmt.Println("BASE_URL: ", BaseURL)
	fmt.Println("TRACE: ", TraceFilename)
	fmt.Println("DUTY: ", DutyCycle)
	fmt.Println("BEGIN: ", BeginAt.Format(time.RFC3339))
	fmt.Println("OUT: ", OutFilename)
	fmt.Println("ERR: ", ErrFilename)
	fmt.Println("---")

	ErrFile, err = os.OpenFile(ErrFilename, os.O_CREATE|os.O_RDWR, 0664)
	if err != nil {
		saveError(err)
		return
	}
	OutFile, err = os.OpenFile(OutFilename, os.O_CREATE|os.O_RDWR, 0664)
	if err != nil || e1 != nil || e2 != nil {
		log.Fatal(err)
	}
	if BeginAt.Unix() < time.Now().Unix() {
		log.Fatal(fmt.Errorf("begin at past time not allowed"))
	}

	wg := sync.WaitGroup{}
	file, err := os.ReadFile(TraceFilename)
	if err != nil {
		log.Fatal(err)
	}
	requests := map[int]*http.Request{}
	data := map[int]map[string]string{}

	for i, line := range strings.Split(string(file), "\n")[1:] {
		row := strings.Split(line, ",")
		if len(row) != 5 {
			continue
		}
		rid := row[0]
		t0, _ := strconv.ParseFloat(row[1], 64)
		fid := row[2]
		dur, _ := strconv.ParseFloat(row[3], 64)
		tb := uint64(dur * 10e9 * DutyCycle)
		ts := uint64(dur * 10e9 * IdleCycle)
		req, err := http.NewRequest("GET", BaseURL, nil)
		if err != nil {
			log.Fatal(err)
		}
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
		query["t0"] = strconv.FormatUint(uint64(t0*10e9), 10)
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
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil || resp.StatusCode != 200 {
				if err == nil {
					err = fmt.Errorf("status %d at RID=%d", resp.StatusCode, rid)
				}
				saveError(err)
				wg.Done()
				return
			}
			go func(query map[string]string, resp *http.Response) {
				ret, err := io.ReadAll(resp.Body)
				if err != nil {
					saveError(err)
					wg.Done()
					return
				}
				retMap := map[string]string{}
				err = json.Unmarshal(ret, &retMap)
				if err != nil {
					saveError(err)
					wg.Done()
					return
				}
				for k, v := range query {
					retMap[k] = v
				}
				retMap["Tf"] = strconv.FormatInt(time.Now().UnixNano(), 10)
				retMapBuf, err := json.Marshal(retMap)
				if err != nil {
					saveError(err)
					wg.Done()
					return
				}
				go saveSuccess(&wg, string(retMapBuf))
			}(row, resp)
		}(i)
	}
	wg.Wait()
}
