package main

import (
	"container/list"
	"fmt"
	"time"
)

func janelaDeslizante(saida chan<- interface{}, entrada <-chan interface{}, tamanho int) {
	buffer := list.New()
	defer close(saida)
	for entrada != nil || buffer.Len() > 0 {
		if buffer.Len() == 0 {
			// nós temos um buffer vazio
			// e um canal de entrada válido
			val := <-entrada
			if val == nil { // assume que nil significa fechado
				entrada = nil // não vai mais ler dados
				continue
			}
			buffer.PushBack(val)
			continue
		}
		select {
		case saida <- buffer.Front().Value:
			// consumidor lê o dado
			buffer.Remove(buffer.Front()) // remove first item
		case val := <-entrada:
			// recebeu nova entrada
			if val == nil {
				// invalida entrada
				entrada = nil
				// continua já que podemos ter dados
				// no buffer
				continue
			}
			if buffer.Len() == tamanho {
				// buffer cheio, descarta dados antigos
				buffer.Remove(buffer.Front())
			}
			// adiciona novo dado no buffer
			buffer.PushBack(val)
		}
	}
}

func leitorLento(in <-chan interface{}) {
	for val := range in {
		fmt.Printf("valor: %v\n", val)
		time.Sleep(4 * time.Second)
	}
}

func sequenciaNumeros(inicial, final int) <-chan interface{} {
	saida := make(chan interface{})
	go func() {
		for i := inicial; i <= final; i++ {
			saida <- i
			time.Sleep(1 * time.Second)
		}
		// após gerar todos os valores, fecha o canal
		close(saida)
	}()
	return saida
}

func main() {
	valores := sequenciaNumeros(1, 10)
	saida := make(chan interface{})
	go leitorLento(saida)
	janelaDeslizante(saida, valores, 3)
}
