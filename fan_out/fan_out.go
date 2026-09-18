package main

import (
	"fmt"
	"sync"
	"time"
)

// trabalhador lê do canal de entrada, que é compartilhado com os demais
// trabalhadores. Cada valor é entregue a exatamente um deles: quem estiver
// livre primeiro, recebe.
func trabalhador(id int, entrada <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for valor := range entrada {
		fmt.Printf("id: %d processando valor: %v\n", id, valor)
		// Simula um processamento demorado
		time.Sleep(100 * time.Millisecond)
	}
}

// fanout distribui os valores de um único canal de entrada entre n
// trabalhadores e só retorna quando todos terminarem.
func fanout(entrada <-chan int, n int) {
	var wg sync.WaitGroup

	wg.Add(n)
	for i := range n {
		go trabalhador(i+1, entrada, &wg)
	}
	wg.Wait()
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
	// Três trabalhadores dividem entre si os dez valores da sequência
	fanout(sequenciaNumeros(1, 10), 3)
}
