package main

import (
	"fmt"
	"time"
)

// janelaDeslizante mantém apenas os `tamanho` itens mais recentes vindos de
// `entrada`, descartando o mais antigo quando a janela enche. Uma única
// goroutine é dona de todo o estado (a fila), então não há disputa entre
// produtor e consumidor pelo buffer.
func janelaDeslizante(entrada <-chan int, saida chan<- int, tamanho int) {
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
		case valor, ok := <-entrada:
			if !ok {
				// Entrada fechada: desabilita este case (canal nil)
				// e continua apenas drenando a fila.
				entrada = nil
				continue
			}
			if len(fila) == tamanho {
				// Janela cheia, descarta o mais antigo e adiciona o novo
				fmt.Printf("Janela Deslizante: Buffer cheio, descartou %v para adicionar %v.\n", fila[0], valor)
				// fila[1:] não libera memória na hora: o array de apoio é
				// mantido até o próximo append realocar. Para uma janela
				// pequena isso é irrelevante, mas é bom saber.
				fila = fila[1:]
			}
			fila = append(fila, valor)

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

func leitorLento(entrada <-chan int, pronto chan<- struct{}) {
	for valor := range entrada {
		fmt.Printf("Consumidor: Recebeu %v\n", valor)
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
	janelaDeslizante(valores, saida, 3)
	<-pronto
	fmt.Println("Fim da execução.")
}
