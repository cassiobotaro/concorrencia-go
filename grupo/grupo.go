package main

import (
	"fmt"
	"sync"
)

// trabalhador processa valores recebidos do canal de entrada e envia resultados para o canal de saída.
// Ele avisa o WaitGroup quando terminar.
func trabalhador(id int, entrada <-chan int, saida chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for valor := range entrada {
		fmt.Printf("id: %d processou valor: %v\n", id, valor)
		saida <- valor * 2
	}

	fmt.Printf("id: %d terminou\n", id)
}

func grupoDeTrabalhadores(entrada <-chan int, nTrabalhadores int) <-chan int {
	saida := make(chan int)
	// Os canais transportam os dados; o WaitGroup apenas conta
	// quantos trabalhadores ainda não terminaram.
	var wg sync.WaitGroup

	// Cria e inicia os trabalhadores
	wg.Add(nTrabalhadores)
	for i := range nTrabalhadores {
		go trabalhador(i+1, entrada, saida, &wg)
	}

	// Goroutine para fechar o canal de saída quando todos os trabalhadores terminarem
	go func() {
		wg.Wait()
		close(saida)
	}()

	return saida
}

func sequenciaNumeros(inicial, final int) <-chan int {
	saida := make(chan int)
	go func() {
		for i := inicial; i <= final; i++ {
			saida <- i
		}
		// Após gerar todos os valores, fecha o canal
		close(saida)
	}()
	return saida
}

func main() {
	// Produz uma sequência de 10 valores
	entrada := sequenciaNumeros(1, 10)
	// Um grupo de trabalhadores irá processar esses números
	saida := grupoDeTrabalhadores(entrada, 2)

	// Somente termina quando todo o trabalho for processado
	for s := range saida {
		fmt.Println(s)
	}
}
