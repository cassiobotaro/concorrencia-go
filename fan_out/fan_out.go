package main

import (
	"fmt"
	"sync"
	"time"
)

func publicar(saida chan<- int, valor int, wg *sync.WaitGroup) {
	timer := time.NewTimer(1 * time.Second)
	// Aguarda 1 segundo ou o canal ser lido
	select {
	case saida <- valor:
	case <-timer.C:
	}
	// Independente do canal ser lido ou não,
	// avisa que a publicação terminou
	wg.Done()
	timer.Stop()
}

func fanout(entrada <-chan int, saidas ...chan<- int) {

	// O agrupamento das publicações é para evitar que
	// o processamento fique bloquando enquanto um canal de saída não é lido
	// e garante que todos os valores serão publicados
	var wg sync.WaitGroup
	for valor := range entrada {
		wg.Add(len(saidas))
		// Publica o valor de entrada em todas as saídas
		for _, saida := range saidas {
			go publicar(saida, valor, &wg)
		}
		wg.Wait()
	}
	// Como a entrada foi consumida, fecha os canais de saída
	for _, saida := range saidas {
		close(saida)
	}
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

func trabalhador(in <-chan int, id int, wg *sync.WaitGroup) {
	for v := range in {
		fmt.Println("id: ", id, " valor: ", v)
	}
	wg.Done()
}

func main() {
	saida1 := make(chan int)
	saida2 := make(chan int)
	// Agrupamos os trabalhadores de forma
	// a aguardar o processamento de todos antes do programa principal
	// ser finalizado
	var wg sync.WaitGroup
	wg.Add(2)
	go trabalhador(saida1, 1, &wg)
	go trabalhador(saida2, 2, &wg)
	fanout(sequenciaNumeros(1, 10), saida1, saida2)
	wg.Wait()
}
