package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// trabalhador lê do canal de entrada, que é compartilhado com os demais
// trabalhadores. Cada valor é entregue a exatamente um deles: quem estiver
// livre primeiro, recebe.
func trabalhador(id int, entrada <-chan int) {
	for valor := range entrada {
		fmt.Printf("id: %d processando valor: %v\n", id, valor)
		// Simula um processamento demorado
		time.Sleep(100 * time.Millisecond)
	}
}

// fanout distribui os valores de um único canal de entrada entre n
// trabalhadores e só retorna quando todos terminarem.
func fanout(entrada <-chan int, n int) {
	var wg sync.WaitGroup

	for i := range n {
		wg.Go(func() {
			trabalhador(i+1, entrada)
		})
	}
	wg.Wait()
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

	// Três trabalhadores dividem entre si os dez valores da sequência
	fanout(sequenciaNumeros(ctx, 1, 10), 3)
}
