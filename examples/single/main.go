package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"syscall"

	"github.com/TykTechnologies/again"
)

func init() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	log.SetPrefix(fmt.Sprintf("pid:%d ", syscall.Getpid()))
}

func main() {

	// Inherit a net.Listener from our parent process or listen anew.
	w, err := again.Listen(func() {})
	if nil != err {
		w = again.New()
		for i := 0; i < 5; i++ {
			// Listen on a TCP or a UNIX domain socket (TCP here).
			l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:400%d", i))
			if nil != err {
				log.Fatalln(err)
			}
			log.Println("listening on", l.Addr())
			if err := w.Listen(fmt.Sprintf("service%d", i), l); err != nil {
				log.Fatalln(err)
			}
			go serve(l)
		}

	} else {
		w.Range(func(s *again.Service) {
			log.Println("listening on", s.Listener.Addr())

			go serve(s.Listener)
		})
		// Kill the parent, now that the child has started successfully.
		if err := again.Kill(); nil != err {
			log.Fatalln(err)
		}
	}

	// Block the main goroutine awaiting signals.
	if _, err := again.Wait(w); nil != err {
		log.Fatalln(err)
	}

	// Do whatever's necessary to ensure a graceful exit like waiting for
	// goroutines to terminate or a channel to become closed.
	//
	// In this case, we'll simply stop listening and wait one second.
	if err := w.Close(); nil != err {
		log.Fatalln(err)
	}
}

// A very rude server that says hello and then closes your connection.
func serve(l net.Listener) {
	http.Serve(l, http.HandlerFunc(serveRequest))
}

func serveRequest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, world!"))
}
