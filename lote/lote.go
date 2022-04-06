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
		buf := make([]req, 0, tamanhoLote)
		valorZero := req{}
		var fechado bool
		// enquanto houver itens para processar
		for !fechado {
			var deveDescarregar bool

			select {
			case r := <-entrada:
				if r == valorZero {
					// close on zero value
					fechado = true
					continue
				}
				// Adiciona o item no buffer
				buf = append(buf, r)
				deveDescarregar = len(buf) == tamanhoLote
			case <-descarga:
				deveDescarregar = true
			}
			if deveDescarregar {
				saida <- buf
				buf = make([]req, 0, tamanhoLote)
			}
		}
		// garante que caso a entrada seja fechada sem preencher o buffer
		// os ultimos itens sejam enviados para processamento
		if len(buf) > 0 {
			saida <- buf
		}
		close(saida)
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
