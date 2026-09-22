package main

import (
	"context"
	"testing"
	"testing/synctest"

	"github.com/cassiobotaro/concorrencia-go/internal/saida"
)

// Com mais de um trabalhador, o id que processa cada valor muda a cada
// execução. O teste usa um único trabalhador para ter uma saída previsível.
func TestExemplo(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		obtida := saida.Capturar(t, func() {
			fanout(sequenciaNumeros(context.Background(), 1, 3), 1)
		})
		saida.Conferir(t, obtida, `
id: 1 processando valor: 1
id: 1 processando valor: 2
id: 1 processando valor: 3
`)
	})
}
