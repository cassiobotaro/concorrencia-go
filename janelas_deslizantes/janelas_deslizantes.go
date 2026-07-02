package main

import (
	"fmt"
	"time"
)

// janelaDeslizante mantém apenas os `tamanho` itens mais recentes vindos de
// `entrada`, descartando o mais antigo quando a janela enche. Uma única
// goroutine é dona de todo o estado (a fila), então não há disputa entre
// produtor e consumidor pelo buffer.
func janelaDeslizante(saida chan<- int, entrada <-chan int, tamanho int) {
	defer close(saida)
	var fila []int

	for entrada != nil || len(fila) > 0 {
		// O case de envio só é habilitado quando há algo na fila:
		// um canal nil nunca é selecionado, o que desabilita o case.
		var envio chan<- int
		var cabeca int
		if len(fila) > 0 {
			envio = saida
			cabeca = fila[0]
		}

		select {
		case val, ok := <-entrada:
			if !ok {
				// Entrada fechada: desabilita este case (canal nil)
				// e continua apenas drenando a fila.
				entrada = nil
				continue
			}
			if len(fila) == tamanho {
				// Janela cheia, descarta o mais antigo e adiciona o novo
				fmt.Printf("Janela Deslizante: Buffer cheio, descartou %v para adicionar %v.\n", fila[0], val)
				fila = fila[1:]
			}
			fila = append(fila, val)

		case envio <- cabeca:
			fmt.Printf("Janela Deslizante: Enviou %v para o consumidor.\n", cabeca)
			fila = fila[1:]
		}
	}
}

// O resto do código permanece o mesmo.
func sequenciaNumeros(inicial, final int) <-chan int {
	saida := make(chan int)
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

func leitorLento(in <-chan int, pronto chan<- struct{}) {
	for val := range in {
		fmt.Printf("Consumidor: Recebeu %v\n", val)
		time.Sleep(4 * time.Second)
	}
	// Fechar o canal é o idioma para sinalizar um evento único
	close(pronto)
}

func main() {
	valores := sequenciaNumeros(1, 10)
	saida := make(chan int)
	pronto := make(chan struct{})
	go leitorLento(saida, pronto)
	janelaDeslizante(saida, valores, 3)
	<-pronto
	fmt.Println("Fim da execução.")
}
