package main

import "fmt"

func dobroFloat(entrada <-chan int) chan<- float64 {
	saida := make(chan float64)
	go func() {
		for valor := range entrada {
			saida <- float64(valor) * 2
		}
		// Após ter terminado de transformar os valores de entrada,
		//  fecha o canal de saida
		close(saida)
	}()
	return saida
}

func sequencia_numeros(inicial, final int) <-chan int {
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
	for valor := range sequencia_numeros(1, 10) {
		fmt.Printf("valor: %v\n", valor)
	}
}
