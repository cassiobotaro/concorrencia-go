package main

import "fmt"

func dobro(entrada <-chan int) <-chan int {
	saida := make(chan int)
	go func() {
		for valor := range entrada {
			saida <- valor * 2
		}
		// Após ter terminado de transformar os valores de entrada,
		//  fecha o canal de saida
		close(saida)
	}()
	return saida
}

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
	for valor := range dobro(dobro(sequenciaNumeros(1, 10))) {
		fmt.Printf("valor: %v\n", valor)
	}
}
