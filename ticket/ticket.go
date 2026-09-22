package main

import (
	"context"
	"fmt"
	"time"
)

type ticket int

func trabalhador(tickets <-chan ticket, trabalhos <-chan func()) {
	for {
		// Lê o trabalho primeiro: se o canal foi fechado, encerra
		// sem gastar um ticket.
		trabalho, ok := <-trabalhos
		if !ok {
			return
		}
		<-tickets // espera autorização antes de executar
		trabalho()
	}
}

// bilheteria emite, no máximo, nTickets por `janela`, ou seja, um ticket a
// cada `janela/nTickets`. O intervalo é contado a partir da entrega, não
// da emissão: se o consumidor demorar a pegar um ticket, o seguinte ainda
// espera o intervalo inteiro, e nunca saem dois de uma vez.
func bilheteria(ctx context.Context, tickets chan<- ticket, janela time.Duration, nTickets int) {
	intervalo := janela / time.Duration(nTickets)
	// Um Timer, e não um Ticker. O Ticker guarda um tick enquanto o envio
	// espera pelo consumidor, e esse tick faria o próximo ticket sair na
	// hora, sem intervalo.
	pausa := time.NewTimer(intervalo)
	defer pausa.Stop()

	for i := 0; ; i++ {
		select {
		case tickets <- ticket(i):
		case <-ctx.Done():
			return
		}

		// espera o intervalo mínimo antes de emitir o próximo ticket
		pausa.Reset(intervalo)
		select {
		case <-pausa.C:
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	tickets := make(chan ticket)
	trabalhos := make(chan func())
	pronto := make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go bilheteria(ctx, tickets, 1*time.Second, 10)
	go func() {
		trabalhador(tickets, trabalhos)
		close(pronto)
	}()

	for i := 0; i <= 30; i++ {
		trabalhos <- func() {
			fmt.Println("processando ticket")
		}
		fmt.Printf("trabalho %d enviado\n", i)
	}

	close(trabalhos)
	<-pronto
}
