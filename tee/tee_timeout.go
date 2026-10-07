package main

import (
	"context"
	"fmt"
	"time"
)

// teeComTimeout é um tee que não espera indefinidamente por um consumidor
// lento: se uma saída não receber o valor dentro de `timeout`, o valor é
// descartado para aquela saída e o tee segue em frente.
// Descartar mensagens é uma decisão de projeto, não parte do padrão.
func teeComTimeout(ctx context.Context, entrada <-chan int, timeout time.Duration, saidas ...chan<- int) {
	// fecha as saídas ao sair, tanto no fim da entrada quanto no cancelamento
	defer func() {
		for _, saida := range saidas {
			close(saida)
		}
	}()
	for valor := range entrada {
		for i, saida := range saidas {
			// Um select por saída: vence o envio, o timeout ou o cancelamento
			select {
			case saida <- valor:
			case <-time.After(timeout):
				// Avisa o descarte em vez de perder o valor em silêncio
				fmt.Printf("tee: descarte por timeout, saida=%d valor=%d\n", i+1, valor)
			case <-ctx.Done():
				return
			}
		}
	}
}
