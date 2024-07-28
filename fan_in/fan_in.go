package main

import (
	"fmt"
)

// fanin combina vários canais de entrada em um único canal de saída.
// Utiliza um canal de sinalização para saber quando todos os canais de entrada foram processados.
func fanin(entradas ...<-chan int) <-chan int {
	saida := make(chan int)

	go func() {
		// Número de canais de entrada
		n := len(entradas)
		// Canal de controle para quando todos os canais de entrada terminarem
		canalTermino := make(chan struct{}, n)

		for _, c := range entradas {
			go func(c <-chan int) {
				for n := range c {
					saida <- n
				}
				// Notifica que este canal foi processado
				canalTermino <- struct{}{}
			}(c)
		}

		// Quando todos os canais de entrada terminarem, fecha o canal de saída
		go func() {
			for i := 0; i < n; i++ {
				<-canalTermino
			}
			close(saida)
		}()
	}()

	return saida
}

// sequenciaNumeros cria um canal que envia uma sequência de números de inicial a final.
func sequenciaNumeros(inicial, final int) <-chan int {
	saida := make(chan int)
	go func() {
		for i := inicial; i <= final; i++ {
			saida <- i
		}
		close(saida)
	}()
	return saida
}

func main() {
	// Combina três canais de sequência em um único canal
	canal := fanin(
		sequenciaNumeros(1, 10),
		sequenciaNumeros(11, 20),
		sequenciaNumeros(21, 30),
	)

	// Lê e imprime os valores do canal combinado
	for valor := range canal {
		fmt.Printf("valor: %v\n", valor)
	}
}
