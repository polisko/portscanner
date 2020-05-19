package main

import (
	"flag"
	"log"
	"sync"
	"time"
)

var lock = sync.RWMutex{}

// A Task in an interface for single worker
type Task interface {
	Process()
	Output()
}

// A Factory is an interface for routine, which creates a Task
type Factory interface {
	Create(host string, port int, resKO map[string]int) Task
}

// Results is a structure, which identifies, how many OK/KO tests went on specific IP
type Results struct {
	OK int
	KO int
}

// Run is a function, which takes factory interface and start the job
func Run(f Factory, workers int, input interface{}, resKO map[string]int) {
	var wg sync.WaitGroup

	in := make(chan Task)
	wg.Add(1)
	go func() {
		switch v := input.(type) {
		case chan Socket:
			for s := range v {
				in <- f.Create(s.IP, s.Port, resKO)
			}
		case []string:
			for _, h := range v {
				in <- f.Create(h, 0, nil)
			}
		}

		close(in)
		wg.Done()
	}()

	out := make(chan Task)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			for t := range in {
				t.Process()
				out <- t
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	for t := range out {
		t.Output()
	}
}

type factoryScan struct {
	timeOut time.Duration
	onlyIP  bool
}

func (f *factoryScan) Create(host string, port int, resKO map[string]int) Task {
	p := &portScan{
		Host:      host,
		Port:      port,
		Timeout:   f.timeOut,
		OnlyIP:    f.onlyIP,
		ResultsKO: resKO,
	}
	return p
}

type factoryHostScan struct {
	timeOut time.Duration
}

func (fh *factoryHostScan) Create(host string, port int, resKO map[string]int) Task {
	hs := &hostScan{
		Host:    host,
		Timeout: fh.timeOut,
		Input:   Input{wellKnown: true},
	}
	return hs
}

func main() {
	var (
		cidr                              string
		startPort, endPort, wrks, timeOut int
		wellKnown, onlyIP                 bool
	)

	flag.StringVar(&cidr, "h", "192.168.1.0/24", "Network address")
	flag.IntVar(&startPort, "s", 20, "Start port")
	flag.IntVar(&endPort, "e", 1024, "End port")
	flag.IntVar(&timeOut, "t", 500, "Timeout for connection attemp")
	flag.IntVar(&wrks, "w", 100, "Number of workers to start parallel")
	flag.BoolVar(&wellKnown, "well", false, "Scan well known ports")
	flag.BoolVar(&onlyIP, "l", false, "Output only IP alivec")

	flag.Parse()

	input := Input{
		cidr:      cidr,
		startPort: startPort,
		endPort:   endPort,
		wellKnown: wellKnown,
	}
	//fmt.Println(input)

	sockets := make(chan Socket)
	go func() {
		err := GenerateSockets(sockets, cidr, input.getPorts())
		if err != nil {
			close(sockets)
			log.Fatal(err)
		}
	}()

	// hosts, err := Hosts(input.cidr)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	f := factoryScan{timeOut: time.Duration(timeOut) * time.Millisecond,
		onlyIP: onlyIP}
	r := make(map[string]int)

	Run(&f, wrks, sockets, r)

}
