package main

import (
	"fmt"
)

// trabalhador processa valores recebidos do canal de entrada e envia resultados para o canal de saída.
// Ele utiliza um canal de sinalização para notificar quando terminar.
func trabalhador(id int, entrada <-chan int, saida chan<- int, terminar chan struct{}) {
	for valor := range entrada {
		fmt.Printf("id: %d processou valor: %v\n", id, valor)
		saida <- valor * 2
	}

	// Envia uma mensagem para o canal de sinalização ao terminar
	fmt.Printf("id: %d terminou\n", id)
	terminar <- struct{}{}
}

func grupoDeTrabalhadores(entrada <-chan int, nTrabalhadores int) chan int {
	saida := make(chan int)
	terminar := make(chan struct{}, nTrabalhadores)

	// Cria e inicia os trabalhadores
	for i := 0; i < nTrabalhadores; i++ {
		go trabalhador(i+1, entrada, saida, terminar)
	}

	// Goroutine para fechar o canal de saída quando todos os trabalhadores terminarem
	go func() {
		// Espera receber sinais de todos os trabalhadores
		for i := 0; i < nTrabalhadores; i++ {
			<-terminar
		}
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
