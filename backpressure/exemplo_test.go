package main

import (
	"fmt"
	"time"
)

// No main, quem enche a fila primeiro depende do relógio. Aqui o consumidor só
// começa a ler depois que o produtor já encheu a fila de 2 posições, então o
// terceiro envio tem de esperar.
func Example() {
	fila := make(chan int, 2)
	go produtor(fila, 3)

	// Dá tempo de o produtor encher a fila e bloquear no terceiro envio
	time.Sleep(100 * time.Millisecond)
	for valor := range fila {
		fmt.Println("Consumidor: leu", valor)
	}

	// Unordered output:
	// Produtor: enviou 1
	// Produtor: enviou 2
	// Produtor: fila cheia, esperando para enviar 3
	// Produtor: enviou 3 após esperar
	// Consumidor: leu 1
	// Consumidor: leu 2
	// Consumidor: leu 3
}
