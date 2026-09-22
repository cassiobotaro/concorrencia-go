package main

import (
	"context"
	"fmt"
	"sync"
)

// trabalhador processa valores recebidos do canal de entrada e envia
// resultados para o canal de saída. O envio disputa com ctx.Done(): se o
// consumidor cancelar, o trabalhador sai em vez de ficar preso no envio.
func trabalhador(ctx context.Context, id int, entrada <-chan int, saida chan<- int) {
	for valor := range entrada {
		fmt.Printf("id: %d processou valor: %v\n", id, valor)
		select {
		case saida <- valor * 2:
		case <-ctx.Done():
			return
		}
	}

	fmt.Printf("id: %d terminou\n", id)
}

func grupoDeTrabalhadores(ctx context.Context, entrada <-chan int, nTrabalhadores int) <-chan int {
	saida := make(chan int)
	// Os canais transportam os dados; o WaitGroup apenas conta
	// quantos trabalhadores ainda não terminaram.
	var wg sync.WaitGroup

	// Cria e inicia os trabalhadores. wg.Go dispara a função em uma nova
	// gorrotina e registra no WaitGroup que ela precisa terminar.
	for i := range nTrabalhadores {
		wg.Go(func() {
			trabalhador(ctx, i+1, entrada, saida)
		})
	}

	// Fecha a saída quando todos os trabalhadores terminarem
	go func() {
		wg.Wait()
		close(saida)
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

	// Produz uma sequência de 10 valores
	entrada := sequenciaNumeros(ctx, 1, 10)
	// Dois trabalhadores dividem esses valores
	saida := grupoDeTrabalhadores(ctx, entrada, 2)

	// Somente termina quando todo o trabalho for processado
	for s := range saida {
		fmt.Println(s)
	}
}
