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

const inputFile = "invokes.csv"
const dutyCycle = .3
const idleCycle = 1 - dutyCycle
const baseURL = "http://192.168.56.9:80/"
const clientId = "7"

var beginAt = time.Date(2024, 2, 23, 9, 30, 0, 0, time.Local)

func saveSuccess(wg *sync.WaitGroup, params map[string]string, response map[string]any) {
	defer wg.Done()
	for k, v := range params {
		response["_"+k] = v
	}

	fmt.Println("[success]: ", response)
}

func saveError(err error) {
	fmt.Println(fmt.Errorf("[error]: %w", err).Error())
}

func main() {
	wg := sync.WaitGroup{}
	file, err := os.ReadFile(inputFile)
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
		tb := uint64(dur * 10e9 * dutyCycle)
		ts := uint64(dur * 10e9 * idleCycle)
		req, err := http.NewRequest("GET", baseURL, nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Add("Content-Type", "text/plain")
		req.Header.Add("Accept", "*/*")
		req.Host = "temul-" + fid + ".default.knative.dev"

		query := map[string]string{}
		data[i] = query
		query["cl"] = clientId
		query["id"] = rid
		query["ts"] = strconv.FormatUint(ts, 10)
		query["tb"] = strconv.FormatUint(tb, 10)
		query["t0"] = strconv.FormatUint(uint64(t0*10e9), 10)
		query["fid"] = fid
		q := req.URL.Query()
		for k, v := range query {
			q.Add(k, v)
		}
		requests[i] = req
	}
	wg.Add(len(requests))
	fmt.Println("Request created.")
	for i := 0; i < len(requests); i++ {
		row := data[i]
		t0, _ := strconv.ParseUint(row["t0"], 10, 64)
		dt := int64(-1)
		for dt = time.Now().Sub(beginAt).Nanoseconds() - int64(t0); dt < 0; dt = time.Now().Sub(beginAt).Nanoseconds() - int64(t0) {
		}
		go func(rid int) {
			fmt.Println(rid, time.Now().Sub(beginAt).Nanoseconds(), dt)
			req := requests[rid]
			row := data[rid]
			row["dt0"] = strconv.FormatInt(dt, 10)
			req.URL.Query().Add("dt0", row["dt0"])
			req.URL.RawQuery = req.URL.Query().Encode()
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
				respMap := map[string]any{}
				err = json.Unmarshal(ret, &respMap)
				if err != nil {
					saveError(err)
					wg.Done()
					return
				}
				go saveSuccess(&wg, query, respMap)
			}(row, resp)
		}(i)
	}
	wg.Wait()
}
