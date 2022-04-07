package main

import (
	"fmt"
	"time"
)

type (
	Trabalho func()
	ticket   int
)

func trabalhador(tickets <-chan ticket, work <-chan Trabalho) {
	for w := range work {
		<-tickets // espera por um ticket
		w()       // executa um trabalho
	}
}

func bilheteria(tickets chan<- ticket, timeout time.Duration, nTickets int) {
	for {
		for i := 0; i < nTickets; i++ {
			tickets <- ticket(i)
		}

		// espera até que mais tickets possam ser emitidos
		<-time.After(timeout)
	}
}

func main() {
	tickets := make(chan ticket)
	trabalhos := make(chan Trabalho)

	go bilheteria(tickets, 1*time.Second, 10)
	go trabalhador(tickets, trabalhos)

	for i := 0; i <= 30; i++ {

		trabalhos <- func() {
			fmt.Println("processando ticket")
		}
		fmt.Println("trabalho ", i, " enviado")
	}

	close(trabalhos)
	close(tickets)
}
