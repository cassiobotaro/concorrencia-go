package main

import (
	"fmt"
	"sync"
)

func fanin(canais_entrada ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	// canal de saída que será compartilhado entre os canais de entrada
	canal_saida := make(chan int)

	// lê os valores de cada canal de entrada e envia para o canal de saída
	// quando todos os valores forem lidos, envia sinal avisando que terminou
	output := func(c <-chan int) {
		for n := range c {
			canal_saida <- n
		}
		// aviso que terminou de ler os valores de um canal
		wg.Done()
	}
	wg.Add(len(canais_entrada))
	// Inicializa uma goroutine de saída para cada canal de entrada em canais_entrada.
	for _, c := range canais_entrada {
		go output(c)
	}

	// Inicia uma goroutine para fechar o canal de saída quando todas as
	// goroutines de entrada terminarem.
	// isto deve ser feito após o wg.Add
	go func() {
		wg.Wait()
		close(canal_saida)
	}()
	return canal_saida
}

func sequencia_numeros(inicial, final int) <-chan int {
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
		sequencia_numeros(1, 10),
		sequencia_numeros(11, 20),
		sequencia_numeros(21, 30),
	)
	for valor := range canal {
		fmt.Printf("valor: %v\n", valor)
	}
}
