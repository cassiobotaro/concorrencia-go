package main

import (
	"context"
	"fmt"
	"sync"
)

// fanin combina vários canais de entrada em um único canal de saída.
// Usa um WaitGroup para saber quando todos os canais de entrada foram
// processados. Cada envio disputa com ctx.Done(), então as gorrotinas saem
// se o consumidor cancelar em vez de ficarem presas no envio.
func fanin(ctx context.Context, entradas ...<-chan int) <-chan int {
	saida := make(chan int)
	var wg sync.WaitGroup

	for _, entrada := range entradas {
		// Uma gorrotina por entrada; o WaitGroup é avisado quando ela termina
		wg.Go(func() {
			for valor := range entrada {
				select {
				case saida <- valor:
				case <-ctx.Done():
					return
				}
			}
		})
	}

	// Quando todos os canais de entrada terminarem, fecha o canal de saída
	go func() {
		wg.Wait()
		close(saida)
	}()

	return saida
}

// sequenciaNumeros envia os inteiros de inicial a final por um canal.
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

	// Combina três canais de sequência em um único canal
	canal := fanin(ctx,
		sequenciaNumeros(ctx, 1, 10),
		sequenciaNumeros(ctx, 11, 20),
		sequenciaNumeros(ctx, 21, 30),
	)

	for valor := range canal {
		fmt.Printf("valor: %v\n", valor)
	}

	// Com um número fixo de entradas, uma única gorrotina com select basta
	// (veja fan_in_select.go)
	canal = faninSelect(ctx,
		sequenciaNumeros(ctx, 31, 40),
		sequenciaNumeros(ctx, 41, 50),
	)
	for valor := range canal {
		fmt.Printf("valor (select): %v\n", valor)
	}
}
