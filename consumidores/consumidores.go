package main

import "fmt"

func trabalhador(canal_entrada <-chan int) {
	for valor := range canal_entrada {
		fmt.Printf("valor: %v\n", valor)
	}
}

func main() {
	entrada := make(chan int)
	// Um trabalhador é iniciado e aguarda por valores no canal de entrada
	go trabalhador(entrada)
	for i := 0; i < 10; i++ {
		entrada <- i
	}
	// Após ter enviado todos os valores, fecha o canal de entrada
	// avisando ao trabalhador que o trabalho terminou
	close(entrada)
}
