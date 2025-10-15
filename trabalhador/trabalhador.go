package main

import "fmt"

func trabalhador(entrada <-chan int) {
	for valor := range entrada {
		fmt.Printf("valor: %v\n", valor)
	}
}

func main() {
	entrada := make(chan int)
	pronto := make(chan struct{})
	// Um trabalhador é iniciado e aguarda por valores no canal de entrada
	go func() {
		trabalhador(entrada)
		pronto <- struct{}{}
	}()
	for i := range 10 {
		entrada <- i
	}
	// Após ter enviado todos os valores, fecha o canal de entrada
	// avisando ao trabalhador que o trabalho terminou
	close(entrada)
	// Aguarda o trabalhador terminar
	<-pronto
}
