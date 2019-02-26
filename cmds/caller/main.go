package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

type ticket struct {
	number uint
}

func main() {

	ticketGenBtn := make(chan interface{})
	ticketPrinter := make(chan *ticket)

	waitingTicket := make(chan *ticket, 128)
	// monitor := make(chan string)

	go func(btn chan interface{}, output chan *ticket) {
		var counter uint
		for {
			select {
			case <-btn:
				counter++
				output <- &ticket{number: counter}
			}
		}
	}(ticketGenBtn, ticketPrinter)

	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/printer/print", func(w http.ResponseWriter, r *http.Request) {
		ticketGenBtn <- nil
		ticket := <-ticketPrinter
		logRec := fmt.Sprintf("ticket %d is printed", ticket.number)
		log.Printf("%s", logRec)
		fmt.Fprintf(w, "number: %d", ticket.number)
		log.Printf("%s", logRec)
		waitingTicket <- ticket
		logRec = fmt.Sprintf("ticket %d is enqueued", ticket.number)
		log.Printf("%s", logRec)
	})

	serverMux.HandleFunc("/monitor", func(w http.ResponseWriter, r *http.Request) {
		// fmt.Fprintf(w, "%s", <-monitor)
	})

	serverMux.HandleFunc("/server/", func(w http.ResponseWriter, r *http.Request) {
		reqCmd := strings.Split(r.URL.String(), "/")

		if len(reqCmd) > 3 && reqCmd[3] == "dequeue" {
			select {
			case ticket := <-waitingTicket:
				logRec := fmt.Sprintf("ticket %d to server %s", ticket.number, reqCmd[2])
				fmt.Fprintf(w, "server: %s<br>", reqCmd[2])
				fmt.Fprintf(w, "number: %d", ticket.number)
				// monitor <- logRec
				log.Printf("%s", logRec)
			case <-r.Cancel:

			}
		} else {
			fmt.Fprintf(w, "404")
			return
		}

	})

	serverMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "404")
	})

	err := http.ListenAndServe(":61234", serverMux)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
