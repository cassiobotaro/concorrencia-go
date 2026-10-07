package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// tee copia cada valor da entrada para todas as saídas: todos os consumidores
// veem todos os valores. O envio é sequencial e sem buffer, então o tee só
// avança quando todas as saídas receberam o valor: um consumidor lento
// atrasa todos os outros. Cada envio disputa com ctx.Done(): se um
// consumidor parar de ler, o tee sai em vez de ficar preso.
func tee(ctx context.Context, entrada <-chan int, saidas ...chan<- int) {
	// fecha as saídas ao sair, tanto no fim da entrada quanto no cancelamento
	defer func() {
		for _, saida := range saidas {
			close(saida)
		}
	}()
	for valor := range entrada {
		for _, saida := range saidas {
			select {
			case saida <- valor:
			case <-ctx.Done():
				return
			}
		}
	}
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

// trabalhador consome os valores de uma das saídas do tee. O parâmetro
// `demora` simula o tempo de processamento de cada valor.
func trabalhador(id int, entrada <-chan int, demora time.Duration) {
	for valor := range entrada {
		fmt.Printf("id: %d valor: %d\n", id, valor)
		time.Sleep(demora)
	}
}

func main() {
	ctx, cancelar := context.WithCancel(context.Background())
	defer cancelar()

	saida1 := make(chan int)
	saida2 := make(chan int)

	// Aguarda o término dos trabalhadores
	var wg sync.WaitGroup
	wg.Go(func() { trabalhador(1, saida1, 0) })
	wg.Go(func() { trabalhador(2, saida2, 0) })

	// Copia a sequência de números para todos os canais de saída
	tee(ctx, sequenciaNumeros(ctx, 1, 10), saida1, saida2)
	wg.Wait()

	// Tee com timeout (veja tee_timeout.go): agora o trabalhador 2 é mais lento
	// do que o timeout, então parte dos valores destinados a ele é descartada.
	saida1 = make(chan int)
	saida2 = make(chan int)

	wg.Go(func() { trabalhador(1, saida1, 0) })
	wg.Go(func() { trabalhador(2, saida2, 250*time.Millisecond) })

	teeComTimeout(ctx, sequenciaNumeros(ctx, 1, 5), 100*time.Millisecond, saida1, saida2)
	wg.Wait()
}
