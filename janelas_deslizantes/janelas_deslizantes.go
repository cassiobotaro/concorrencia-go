package main

import (
	"fmt"
	"time"
)

func janelaDeslizante(saida chan<- interface{}, entrada <-chan interface{}, tamanho int) {
	buffer := make(chan interface{}, tamanho)
	defer close(saida)

	// Lógica de leitura do produtor
	go func() {
		defer close(buffer)
		for val := range entrada {
			// Se o buffer estiver cheio, descarta o mais antigo
			if len(buffer) == tamanho {
				<-buffer
				fmt.Printf("Janela Deslizante: Buffer cheio, descartou valor antigo para adicionar %v.\n", val)
			}
			buffer <- val
		}
	}()

	// Lógica de envio para o consumidor
	for val := range buffer {
		saida <- val
		fmt.Printf("Janela Deslizante: Enviou %v para o consumidor.\n", val)
	}
}

// O resto do código permanece o mesmo.
func sequenciaNumeros(inicial, final int) <-chan interface{} {
	saida := make(chan interface{})
	go func() {
		for i := inicial; i <= final; i++ {
			saida <- i
			fmt.Printf("Produtor: Enviou %d\n", i)
			time.Sleep(1 * time.Second)
		}
		close(saida)
	}()
	return saida
}

func leitorLento(in <-chan interface{}) {
	for val := range in {
		fmt.Printf("Consumidor: Recebeu %v\n", val)
		time.Sleep(4 * time.Second)
	}
}

func main() {
	valores := sequenciaNumeros(1, 10)
	saida := make(chan interface{})
	go leitorLento(saida)
	janelaDeslizante(saida, valores, 3)
	fmt.Println("Fim da execução.")
}
