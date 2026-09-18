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

func trabalhador(tickets <-chan ticket, trabalhos <-chan Trabalho) {
	for {
		// Lê o trabalho primeiro: se o canal foi fechado, encerra
		// sem gastar um ticket.
		trabalho, ok := <-trabalhos
		if !ok {
			return // canal de trabalhos fechado
		}
		<-tickets  // espera autorização antes de executar
		trabalho() // executa um trabalho
	}
}

// bilheteria emite, no máximo, nTickets por intervalo `timeout`,
// ou seja, um ticket a cada `timeout/nTickets`. Garante o teto mesmo com consumidor
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
		// Sinaliza o término fechando o canal
		close(pronto)
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
