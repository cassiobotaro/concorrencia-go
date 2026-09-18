package main

import (
	"fmt"
	"sync"
	"time"
)

// tee copia cada valor da entrada para todas as saídas: todos os consumidores
// veem todos os valores. O envio é sequencial e sem buffer, então o tee só
// avança quando todas as saídas receberam o valor: um consumidor lento
// atrasa todos os outros.
func tee(entrada <-chan int, saidas ...chan<- int) {
	for valor := range entrada {
		for _, saida := range saidas {
			saida <- valor
		}
	}
	// Como a entrada foi consumida, fecha os canais de saída
	for _, saida := range saidas {
		close(saida)
	}
}

func sequenciaNumeros(inicial, final int) <-chan int {
	saida := make(chan int)
	go func() {
		for i := inicial; i <= final; i++ {
			saida <- i
		}
		// após gerar todos os valores, fecha o canal
		close(saida)
	}()
	return saida
}

// trabalhador consome os valores de uma das saídas do tee. O parâmetro
// `demora` simula o tempo de processamento de cada valor.
func trabalhador(id int, entrada <-chan int, demora time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()
	for valor := range entrada {
		fmt.Println("id: ", id, " valor: ", valor)
		time.Sleep(demora)
	}
}

func main() {
	saida1 := make(chan int)
	saida2 := make(chan int)

	// Aguarda o término dos trabalhadores
	var wg sync.WaitGroup
	wg.Add(2)
	go trabalhador(1, saida1, 0, &wg)
	go trabalhador(2, saida2, 0, &wg)

	// Copia a sequência de números para todos os canais de saída
	tee(sequenciaNumeros(1, 10), saida1, saida2)
	wg.Wait()

	// Tee com timeout (veja tee_timeout.go): agora o trabalhador 2 é mais lento
	// do que o timeout, então parte dos valores destinados a ele é descartada.
	saida1 = make(chan int)
	saida2 = make(chan int)

	wg.Add(2)
	go trabalhador(1, saida1, 0, &wg)
	go trabalhador(2, saida2, 250*time.Millisecond, &wg)

	teeComTimeout(sequenciaNumeros(1, 5), 100*time.Millisecond, saida1, saida2)
	wg.Wait()
}
