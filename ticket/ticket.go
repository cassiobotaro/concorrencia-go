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
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()
	for {
		for i := range nTickets {
			tickets <- ticket(i)
		}

		// espera até que mais tickets possam ser emitidos
		<-ticker.C
	}
}

func main() {
	tickets := make(chan ticket)
	trabalhos := make(chan Trabalho)
	pronto := make(chan struct{})

	go bilheteria(tickets, 1*time.Second, 10)
	go func() {
		trabalhador(tickets, trabalhos)
		pronto <- struct{}{}
	}()

	for i := 0; i <= 30; i++ {
		trabalhos <- func() {
			fmt.Println("processando ticket")
		}
		fmt.Println("trabalho ", i, " enviado")
	}

	close(trabalhos)
	<-pronto
}
