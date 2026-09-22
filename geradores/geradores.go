package main

import (
	"context"
	"fmt"
)

// sequenciaNumeros gera os inteiros de inicial a final em uma gorrotina e
// os envia por um canal. Cada envio disputa com ctx.Done(): se o consumidor
// cancelar o contexto, a gorrotina sai em vez de ficar presa no envio.
func sequenciaNumeros(ctx context.Context, inicial, final int) <-chan int {
	saida := make(chan int)
	go func() {
		// fecha o canal ao sair, tanto no fim quanto no cancelamento
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

	for valor := range sequenciaNumeros(ctx, 1, 1000) {
		fmt.Printf("valor: %v\n", valor)
	}
}
