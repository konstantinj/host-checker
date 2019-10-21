package main

import (
	"bufio"
	"flag"
	"log"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var errLog *log.Logger
var okLog *log.Logger

// call like this `go run main.go < hosts.txt 2>> offline.txt 1>> online.txt`
func main() {
	ports := flag.String("ports", "20,21,22,23,25,53,80,143,389,443,1194,1337,2003,2004,3000,3306,3389,5000,5672,5800,5900,6379,8000,8080,8091,8443,9000,9200,9300,11211,15672,27017", "which ports to check")
	procs := flag.String("procs", "0", "how much concurrency")
	flag.Parse()

	errLog = log.New(os.Stderr, "", 0)
	okLog = log.New(os.Stdout, "", 0)

	portList := strings.Fields(strings.Replace(*ports, ",", " ", -1))
	procsN, _ := strconv.Atoi(*procs)

	if procsN == 0 {
		procsN = math.MaxInt16
	}

	s := bufio.NewScanner(os.Stdin)
	var wg sync.WaitGroup
	i := 0

	for s.Scan() {
		host := s.Text()

		if len(strings.Fields(s.Text())) > 0 {
			host = strings.Fields(s.Text())[0]
			domainType := strings.Fields(s.Text())[1]
			if domainType != "A" && domainType != "CNAME" {
				continue
			}
		}

		go rawConnect(host, portList, &wg)

		i++
		if i == procsN {
			wg.Wait()
			i = 0
		}
	}
	wg.Wait()
}

func rawConnect(host string, ports []string, wg *sync.WaitGroup) {
	wg.Add(1)
	portErrors := 0
	for _, network := range []string{"tcp", "udp"} {
		for _, port := range ports {
			timeout := time.Second
			conn, err := net.DialTimeout(network, net.JoinHostPort(host, port), timeout)
			if err != nil {
				portErrors++
			}
			if conn != nil {
				defer conn.Close()
				okLog.Println(host)
				wg.Done()
				return
			}
			if portErrors == 2*len(ports) {
				errLog.Println(host)
				wg.Done()
				return
			}
		}
	}
	wg.Done()
}
