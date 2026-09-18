package main

import (
	"context"
	"fmt"
	"time"
)

// O main leva 3 segundos. Aqui a mesma montagem roda com 5 tickets a cada
// 50ms (um a cada 10ms): três trabalhos precisam de, no mínimo, dois
// intervalos, pois só o primeiro ticket sai de imediato.
func Example() {
	tickets := make(chan ticket)
	trabalhos := make(chan Trabalho)
	pronto := make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go bilheteria(ctx, tickets, 50*time.Millisecond, 5)
	go func() {
		trabalhador(tickets, trabalhos)
		close(pronto)
	}()

	inicio := time.Now()
	for i := range 3 {
		trabalhos <- func() {
			fmt.Println("processando trabalho", i)
		}
	}
	close(trabalhos)
	<-pronto

	fmt.Println("respeitou a taxa:", time.Since(inicio) >= 20*time.Millisecond)

	// Output:
	// processando trabalho 0
	// processando trabalho 1
	// processando trabalho 2
	// respeitou a taxa: true
}
