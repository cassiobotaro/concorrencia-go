package main

import "fmt"

func sequenciaNumeros(inicial, final int) <-chan int {
	saida := make(chan int)
	go func() {
		for i := inicial; i <= final; i++ {
			saida <- i
		}
		// após gerar todos os valores, fecha o canal
		close(saida)
	}()
	return saida
}

func main() {
	valores := sequenciaNumeros(1, 1000)
	for valor := range valores {
		fmt.Printf("valor: %v\n", valor)
	}
}
