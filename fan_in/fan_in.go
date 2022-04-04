package main

import (
	"fmt"
	"sync"
)

func fanin(canaisEntrada ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	// canal de saída que será compartilhado entre os canais de entrada
	canalSaida := make(chan int)

	// lê os valores de cada canal de entrada e envia para o canal de saída
	// quando todos os valores forem lidos, envia sinal avisando que terminou
	output := func(c <-chan int) {
		for n := range c {
			canalSaida <- n
		}
		// aviso que terminou de ler os valores de um canal
		wg.Done()
	}
	wg.Add(len(canaisEntrada))
	// Inicializa uma goroutine de saída para cada canal de entrada em canais_entrada.
	for _, c := range canaisEntrada {
		go output(c)
	}

	// Inicia uma goroutine para fechar o canal de saída quando todas as
	// goroutines de entrada terminarem.
	// isto deve ser feito após o wg.Add
	go func() {
		wg.Wait()
		close(canalSaida)
	}()
	return canalSaida
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

func main() {
	canal := fanin(
		sequenciaNumeros(1, 10),
		sequenciaNumeros(11, 20),
		sequenciaNumeros(21, 30),
	)
	for valor := range canal {
		fmt.Printf("valor: %v\n", valor)
	}
}
