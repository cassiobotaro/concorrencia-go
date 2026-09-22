package main

import (
	"fmt"
	"time"
)

type req struct {
	valor int
}

func processar(lote []req) {
	fmt.Printf("processando lote com valores: %v\n", lote)
}

func processadorLotes(entrada <-chan []req) <-chan struct{} {
	pronto := make(chan struct{})
	go func() {
		for lote := range entrada {
			processar(lote)
		}
		// Sinaliza o término do processamento fechando o canal
		close(pronto)
	}()
	return pronto
}

// processamentoLotes agrupa os itens da entrada em lotes. Um lote é enviado
// quando enche (`tamanhoLote`), quando passa o `intervalo` sem que ele tenha
// enchido, ou quando chega um sinal manual pelo canal `descarga`.
func processamentoLotes(entrada <-chan req, descarga <-chan struct{}, tamanhoLote int, intervalo time.Duration) <-chan []req {
	saida := make(chan []req)
	go func() {
		defer close(saida)
		buf := make([]req, 0, tamanhoLote)

		// O prazo conta a partir do primeiro item de cada lote: o Timer é
		// armado quando o lote começa e parado quando ele sai. Um Ticker
		// marcaria o tempo por conta própria, e um lote começado logo antes
		// do tick sairia quase vazio. Nasce parado porque ainda não há lote.
		prazo := time.NewTimer(intervalo)
		prazo.Stop()

		// descarregar envia o lote atual, se houver algo nele
		descarregar := func() {
			if len(buf) == 0 {
				return
			}
			prazo.Stop()
			saida <- buf
			// Um novo slice é criado em vez de reaproveitar com buf[:0]:
			// o consumidor pode ainda estar lendo o lote enviado, e reutilizar
			// o mesmo array de apoio sobrescreveria dados em uso (aliasing).
			buf = make([]req, 0, tamanhoLote)
		}

		for {
			select {
			// enquanto houver itens para processar
			case item, ok := <-entrada:
				if !ok {
					// envia o que tiver no buffer antes de sair
					descarregar()
					// para o loop quando o canal de entrada for fechado
					return
				}
				// Primeiro item do lote: começa a contar o prazo
				if len(buf) == 0 {
					prazo.Reset(intervalo)
				}
				buf = append(buf, item)
				// se o buffer estiver cheio, descarrega
				if len(buf) == tamanhoLote {
					descarregar()
				}

			// Se o intervalo passou, descarrega o que tiver no buffer
			case <-prazo.C:
				descarregar()

			// Se receber um sinal de descarga, descarrega o que tiver no buffer
			case <-descarga:
				descarregar()
			}
		}
	}()
	return saida
}

func main() {
	entrada := make(chan req)
	descarga := make(chan struct{})

	// inicia de forma concorrente o processamento em lotes:
	// lotes de 3 itens ou 100ms, o que acontecer primeiro
	saida := processamentoLotes(entrada, descarga, 3, 100*time.Millisecond)
	// O consumidor de lotes será iniciado de forma concorrente
	pronto := processadorLotes(saida)

	entrada <- req{valor: 1}
	entrada <- req{valor: 2}
	entrada <- req{valor: 3}

	// Envia mais dois itens e força a descarga do lote
	// pelo canal de descarga
	entrada <- req{valor: 4}
	entrada <- req{valor: 5}
	descarga <- struct{}{}

	// Envia um item e espera: o lote não enche, mas o intervalo
	// de 100ms passa e ele é descarregado mesmo assim
	entrada <- req{valor: 6}
	time.Sleep(150 * time.Millisecond)

	// Envia mais dois itens, não o suficiente para descarregar
	// o lote.
	entrada <- req{valor: 7}
	entrada <- req{valor: 8}
	// Eles serão processados mesmo assim.

	close(entrada)

	// Aguarda todo o processamento do processador de lotes
	// antes de encerrar o programa
	<-pronto
}
