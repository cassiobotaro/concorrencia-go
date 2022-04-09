package main

import (
	"fmt"
	"sync"
)

func trabalhador(id int, entrada <-chan int, saida chan<- int, grupo *sync.WaitGroup) {
	for valor := range entrada {
		fmt.Printf("id: %d processou valor: %v\n", id, valor)
		saida <- valor * 2
	}
	grupo.Done()
}

func grupoDeTrabalhadores(entrada <-chan int, nTrabalhadores int) chan int {
	saida := make(chan int)
	var wg sync.WaitGroup
	for i := 0; i < nTrabalhadores; i++ {
		go trabalhador(i+1, entrada, saida, &wg)
	}
	wg.Add(nTrabalhadores)
	go func() {
		// Quando todos os trabalhadores estiverem terminado
		// informa que o grupo não vai mais enviar resultados
		wg.Wait()
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
	// Produz uma sequência de 10 valores
	entrada := sequenciaNumeros(1, 10)
	// Um grupo de trabalhadores irá processar esses números
	saida := grupoDeTrabalhadores(entrada, 2)

	// somente termina quando todo o trabalho for processado
	for s := range saida {
		fmt.Println(s)
	}
}
