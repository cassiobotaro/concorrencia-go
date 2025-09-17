package main

import (
	"fmt"
)

type req struct {
	valor int
}

func processar(lote []req) {
	fmt.Println("processando lote com valores: ", lote)
}

func processadorLotes(entrada <-chan []req) chan bool {
	pronto := make(chan bool)
	go func() {
		for lote := range entrada {
			processar(lote)
		}
		pronto <- true
	}()
	return pronto
}

func processamentoLotes(entrada <-chan req, descarga <-chan struct{}, tamanhoLote int) chan []req {
	saida := make(chan []req)
	go func() {
		defer close(saida)
		buf := make([]req, 0, tamanhoLote)

		for {
			select {
			// enquanto houver itens para processar
			case r, ok := <-entrada:
				if !ok {
					// envia o que tiver no buffer antes de sair
					if len(buf) > 0 {
						saida <- buf
					}
					// para o loop quando o canal de entrada for fechado
					return
				}
				// Adiciona o item no buffer
				buf = append(buf, r)
				// se o buffer estiver cheio, descarrega
				if len(buf) == tamanhoLote {
					saida <- buf
					buf = make([]req, 0, tamanhoLote)
				}

			// Se receber um sinal de descarga, descarrega o que tiver no buffer
			case <-descarga:
				if len(buf) > 0 {
					saida <- buf
					buf = make([]req, 0, tamanhoLote)
				}
			}
		}
	}()
	return saida
}

func main() {
	entrada := make(chan req)
	descarga := make(chan struct{})

	// inicia de forma concorrente o processamento em lotes
	saida := processamentoLotes(entrada, descarga, 3)
	// O consumidor de lotes será iniciado de forma concorrente
	pronto := processadorLotes(saida)

	entrada <- req{valor: 1}
	entrada <- req{valor: 2}
	entrada <- req{valor: 3}

	// Envia mais dois itens, porém força a descarga
	// através de um sinal
	entrada <- req{valor: 4}
	entrada <- req{valor: 5}
	descarga <- struct{}{}

	// Envia mais dois itens, não o suficiente para descarregar
	// o lote.
	entrada <- req{valor: 6}
	entrada <- req{valor: 7}
	// Eles serão processados mesmo assim.

	close(entrada)

	// Aguarda todo o processamento do processador de lotes
	// antes de encerrar o programa
	<-pronto
}
