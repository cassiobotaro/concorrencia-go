package main

import (
	"context"
	"fmt"
)

// dobro é um estágio: lê da entrada, escreve o dobro na saída e fecha a
// saída quando a entrada acaba. O select em cada envio deixa o estágio sair
// quando o contexto é cancelado, em vez de ficar preso esperando um leitor.
func dobro(ctx context.Context, entrada <-chan int) <-chan int {
	saida := make(chan int)
	go func() {
		defer close(saida)
		for valor := range entrada {
			select {
			case saida <- valor * 2:
			case <-ctx.Done():
				return
			}
		}
	}()
	return saida
}

func sequenciaNumeros(ctx context.Context, inicial, final int) <-chan int {
	saida := make(chan int)
	go func() {
		defer close(saida)
		for i := inicial; i <= final; i++ {
			select {
			case saida <- i:
			case <-ctx.Done():
				return
			}
		}
	}()
	return saida
}

func main() {
	ctx, cancelar := context.WithCancel(context.Background())
	defer cancelar()

	// O mesmo contexto atravessa o gerador e os dois estágios
	for valor := range dobro(ctx, dobro(ctx, sequenciaNumeros(ctx, 1, 10))) {
		fmt.Printf("valor: %v\n", valor)
	}
}
