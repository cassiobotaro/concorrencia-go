package main

import (
	"context"
	"fmt"
	"time"
)

type (
	Trabalho func()
	ticket   int
)

func trabalhador(tickets <-chan ticket, work <-chan Trabalho) {
	for {
		<-tickets // espera autorização antes de consumir trabalho
		w, ok := <-work
		if !ok {
			return // canal de trabalhos fechado
		}
		w() // executa um trabalho
	}
}

func bilheteria(ctx context.Context, tickets chan<- ticket, timeout time.Duration, nTickets int) {
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()
	for {
		for i := range nTickets {
			select {
			case tickets <- ticket(i):
			case <-ctx.Done():
				return
			}
		}

		// espera até que mais tickets possam ser emitidos
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	tickets := make(chan ticket)
	trabalhos := make(chan Trabalho)
	pronto := make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go bilheteria(ctx, tickets, 1*time.Second, 10)
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
	cancel()
}
