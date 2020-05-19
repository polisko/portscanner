package main

import (
	"fmt"
	"net"
	"time"
)

type portScan struct {
	Host      string
	Port      int
	Timeout   time.Duration
	Connect   bool
	ResultsKO map[string]int
	OnlyIP    bool
}

func (ps *portScan) Process() {
	//t1 := time.Now()
	// d := net.Dialer{Timeout: ps.Timeout}
	// ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	// defer cancel()
	// lock.RLock()
	// if ps.ResultsKO[ps.Host] > 50 {
	// 	// lock.RUnlock()
	// 	// return
	// }
	// lock.RUnlock()

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ps.Host, ps.Port), ps.Timeout)

	//fmt.Println("connecting ", fmt.Sprintf("%s:%d", ps.Host, ps.Port))
	//conn, err := d.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", ps.Host, ps.Port))
	if err == nil {
		conn.Close()
		ps.Connect = true

	} else {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			lock.Lock()
			ps.ResultsKO[ps.Host]++
			lock.Unlock()
		}
	}
	//fmt.Println(time.Since(t1))
}

func (ps *portScan) Output() {

	st := "closed"
	if ps.Connect {
		st = "opened"
		if ps.OnlyIP {
			fmt.Printf("%s\n", ps.Host)
		} else {
			fmt.Printf("Host: %s, port %d: %s\n", ps.Host, ps.Port, st)
		}
	} else {
		// lock.RLock()
		// fmt.Printf("Host: %s, port %d: %v\n", ps.Host, ps.Port, ps.ResultsKO[ps.Host])
		// lock.RUnlock()
	}
}

type hostScan struct {
	Host      string
	Input     Input
	Timeout   time.Duration
	Alive     bool
	AlivePort int
}

func (hs *hostScan) Process() {
	timeOutsErr := 0
	ports := hs.Input.getPorts()
	for _, v := range ports {
		//fmt.Printf("%s:%d errors:%d\n", hs.Host, v, timeOutsErr)
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", hs.Host, v), hs.Timeout)
		if err == nil {
			conn.Close()
			hs.Alive = true
			hs.AlivePort = v
			break
		} else {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				timeOutsErr++
			}
		}
		if timeOutsErr > int(len(ports)/10) {
			hs.Alive = false
			hs.AlivePort = int(len(ports) / 10)
			break
		}
	}
}

func (hs *hostScan) Output() {
	if hs.Alive {
		fmt.Printf("Host %s is alive, responded at port %d\n", hs.Host, hs.AlivePort)
	} else {
		fmt.Printf("Host %s is down, responded with timeout on more than %d ports\n", hs.Host, hs.AlivePort)

	}
}
