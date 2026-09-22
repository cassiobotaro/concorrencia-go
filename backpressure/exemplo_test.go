package main

import (
	"fmt"
	"testing"
	"testing/synctest"

	"github.com/cassiobotaro/concorrencia-go/internal/saida"
)

// No main, quem enche a fila primeiro depende do relógio. Aqui o consumidor só
// começa a ler depois que o produtor já encheu a fila de 2 posições, então o
// terceiro envio tem de esperar.
func TestExemplo(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		obtida := saida.Capturar(t, func() {
			fila := make(chan int, 2)
			go produtor(fila, 3)

			// Espera até o produtor ficar bloqueado no terceiro envio
			synctest.Wait()
			for valor := range fila {
				fmt.Println("Consumidor: leu", valor)
			}
		})
		saida.ConferirSemOrdem(t, obtida, `
Produtor: enviou 1
Produtor: enviou 2
Produtor: fila cheia, esperando para enviar 3
Produtor: enviou 3 após esperar
Consumidor: leu 1
Consumidor: leu 2
Consumidor: leu 3
`)
	})
}
