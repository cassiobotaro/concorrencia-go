package main

import (
	"fmt"
	"time"
)

// teeComTimeout é um tee que não espera indefinidamente por um consumidor
// lento: se uma saída não receber o valor dentro de `timeout`, o valor é
// descartado para aquela saída e o tee segue em frente.
// Descartar mensagens é uma decisão de projeto, não parte do padrão.
func teeComTimeout(entrada <-chan int, timeout time.Duration, saidas ...chan<- int) {
	for valor := range entrada {
		for i, saida := range saidas {
			// Um select por saída: vence o envio ou o timeout
			select {
			case saida <- valor:
			case <-time.After(timeout):
				// Avisa o descarte em vez de perder o valor em silêncio
				fmt.Printf("tee: descarte por timeout, saida=%d valor=%d\n", i+1, valor)
			}
		}
	}
	// Como a entrada foi consumida, fecha os canais de saída
	for _, saida := range saidas {
		close(saida)
	}
}
