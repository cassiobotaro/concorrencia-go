package main

import (
	"fmt"
	"time"
)

func janelaDeslizante(saida chan<- any, entrada <-chan any, tamanho int) {
	buffer := make(chan any, tamanho)
	defer close(saida)

	// Lógica de leitura do produtor
	go func() {
		defer close(buffer)
		for val := range entrada {
			// Tenta enviar para o buffer
			select {
			case buffer <- val:
				// Enviou com sucesso
			default:
				// Buffer cheio, descarta o mais antigo e adiciona o novo
				<-buffer
				fmt.Printf("Janela Deslizante: Buffer cheio, descartou valor antigo para adicionar %v.\n", val)
				buffer <- val
			}
		}
	}()

	// Lógica de envio para o consumidor
	for val := range buffer {
		saida <- val
		fmt.Printf("Janela Deslizante: Enviou %v para o consumidor.\n", val)
	}
}

// O resto do código permanece o mesmo.
func sequenciaNumeros(inicial, final int) <-chan any {
	saida := make(chan any)
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

func leitorLento(in <-chan any, pronto chan<- struct{}) {
	for val := range in {
		fmt.Printf("Consumidor: Recebeu %v\n", val)
		time.Sleep(4 * time.Second)
	}
	pronto <- struct{}{}
}

func main() {
	valores := sequenciaNumeros(1, 10)
	saida := make(chan any)
	pronto := make(chan struct{})
	go leitorLento(saida, pronto)
	janelaDeslizante(saida, valores, 3)
	<-pronto
	fmt.Println("Fim da execução.")
}
