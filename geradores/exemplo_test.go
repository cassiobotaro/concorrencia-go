package main

import (
	"context"
	"fmt"
)

func Example() {
	ctx, cancelar := context.WithCancel(context.Background())
	for valor := range sequenciaNumeros(ctx, 1, 3) {
		fmt.Printf("valor: %v\n", valor)
	}

	// Parando antes do fim: cancela e drena até o canal fechar, o que
	// prova que a gorrotina do gerador terminou.
	valores := sequenciaNumeros(ctx, 1, 1000)
	fmt.Printf("valor: %v\n", <-valores)
	cancelar()
	for range valores {
	}
	fmt.Println("gerador encerrado")

	// Output:
	// valor: 1
	// valor: 2
	// valor: 3
	// valor: 1
	// gerador encerrado
}
