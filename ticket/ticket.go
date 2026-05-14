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

// bilheteria emite, no máximo, nTickets por intervalo `timeout` —
// um ticket a cada `timeout/nTickets`. Garante o teto mesmo com consumidor
// lento, em troca de não permitir rajadas (nenhuma janela "extra" no início).
func bilheteria(ctx context.Context, tickets chan<- ticket, timeout time.Duration, nTickets int) {
	intervalo := timeout / time.Duration(nTickets)
	ticker := time.NewTicker(intervalo)
	defer ticker.Stop()

	var i int
	for {
		select {
		case tickets <- ticket(i):
			i++
		case <-ctx.Done():
			return
		}

		// espera o intervalo mínimo antes de emitir o próximo ticket
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
