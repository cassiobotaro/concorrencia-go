package main

import "fmt"

// elo recebe um valor da vizinha da direita, soma 1 e passa para a esquerda.
func elo(esquerda chan<- int, direita <-chan int) {
	esquerda <- 1 + <-direita
}

func main() {
	const n = 10000

	// Monta a corrente da esquerda para a direita: cada gorrotina fica
	// bloqueada esperando o valor da vizinha.
	pontaEsquerda := make(chan int)
	esquerda := pontaEsquerda
	var direita chan int
	for range n {
		direita = make(chan int)
		go elo(esquerda, direita)
		esquerda = direita
	}

	// Solta o primeiro valor na ponta direita...
	go func() { direita <- 1 }()
	// ...e espera ele atravessar as 10 mil gorrotinas.
	fmt.Println(<-pontaEsquerda)
}
