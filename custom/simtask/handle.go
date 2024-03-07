package function

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
	"golang.org/x/sys/unix"
	"net/http"
	"strconv"
	"time"
)

const Version = "0.1.1"

func Handle(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
	_, _ = cpu.Percent(0, true)
	_, _ = disk.IOCounters()
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

	res := map[string]any{}
	res["rt0"] = strconv.FormatInt(rt0.UnixNano(), 10)
	res["rtb"] = strconv.FormatInt(rtb.Nanoseconds(), 10)
	res["rit"] = strconv.FormatInt(rit, 10)
	res["rts"] = strconv.FormatInt(rts.Nanoseconds(), 10)
	res["rdt"] = strconv.FormatInt(rdt.Nanoseconds(), 10)
	res["rtf"] = strconv.FormatInt(rtf.UnixNano(), 10)

	times, err := cpu.Times(true)
	if err == nil {
		res["cpu_times"] = times
	}
	cpupc, err := cpu.Percent(0, true)
	if err == nil {
		res["cpu_pc"] = cpupc
	}
	iouse, err := disk.IOCounters("/dev/xvda1", "/dev/xvda2", "/dev/sda") //sda for local tests
	if err == nil {
		res["iouse"] = iouse
	}
	memstat, err := mem.VirtualMemory()
	if err == nil {
		res["memstat"] = memstat
	}
	memexstat, err := mem.VirtualMemory()
	if err == nil {
		res["memexstat"] = memexstat
	}
	avgstat, err := load.Avg()
	if err == nil {
		res["load"] = avgstat
	}
	miscstat, err := load.Avg()
	if err == nil {
		res["miscstat"] = miscstat
	}
	proc, err := process.NewProcess(int32(unix.Getpid()))
	psconn, err := proc.Connections()
	if err == nil {
		res["psconn"] = psconn
	}
	psio, err := proc.IOCounters()
	if err == nil {
		res["psio"] = psio
	}
	psmem, err := proc.MemoryInfo()
	if err == nil {
		res["psmem"] = psmem
	}
	pstimes, err := proc.Times()
	if err == nil {
		res["pstimes"] = pstimes
	}
	pscpupc, err := proc.CPUPercent()
	if err == nil {
		res["pscpupc"] = pscpupc
	}
	psmempc, err := proc.MemoryPercent()
	if err == nil {
		res["psmempc"] = psmempc
	}
	pccreatets, err := proc.CreateTime()
	if err == nil {
		res["pccreatets"] = pccreatets
	}
	psctxsw, err := proc.NumCtxSwitches()
	if err == nil {
		res["psctxsw"] = psctxsw
	}
	psnfd, err := proc.NumFDs()
	if err == nil {
		res["psnfd"] = psnfd
	}
	psth, err := proc.Threads()
	if err == nil {
		res["psth"] = psth
	}

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
