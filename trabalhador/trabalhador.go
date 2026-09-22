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
	go func() {
		trabalhador(entrada)
		// Fechar o canal é o idioma para sinalizar um evento único:
		// comunica "terminou" a qualquer número de leitores.
		close(pronto)
	}()
	for i := range 10 {
		entrada <- i
	}
	// Fechar a entrada encerra o range do trabalhador
	close(entrada)
	<-pronto
}
