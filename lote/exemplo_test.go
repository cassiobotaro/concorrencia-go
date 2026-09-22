package main

import (
	"testing"
	"testing/synctest"

	"github.com/cassiobotaro/concorrencia-go/internal/saida"
)

func TestExemplo(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		saida.Conferir(t, saida.Capturar(t, main), `
processando lote com valores: [{1} {2} {3}]
processando lote com valores: [{4} {5}]
processando lote com valores: [{6}]
processando lote com valores: [{7} {8}]
`)
	})
}
