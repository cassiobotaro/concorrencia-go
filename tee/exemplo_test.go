package main

import (
	"testing"
	"testing/synctest"

	"github.com/cassiobotaro/concorrencia-go/internal/saida"
)

func TestExemplo(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		saida.ConferirSemOrdem(t, saida.Capturar(t, main), `
id: 1 valor: 1
id: 1 valor: 1
id: 1 valor: 10
id: 1 valor: 2
id: 1 valor: 2
id: 1 valor: 3
id: 1 valor: 3
id: 1 valor: 4
id: 1 valor: 4
id: 1 valor: 5
id: 1 valor: 5
id: 1 valor: 6
id: 1 valor: 7
id: 1 valor: 8
id: 1 valor: 9
id: 2 valor: 1
id: 2 valor: 1
id: 2 valor: 10
id: 2 valor: 2
id: 2 valor: 3
id: 2 valor: 4
id: 2 valor: 4
id: 2 valor: 5
id: 2 valor: 6
id: 2 valor: 7
id: 2 valor: 8
id: 2 valor: 9
tee: descarte por timeout, saida=2 valor=2
tee: descarte por timeout, saida=2 valor=3
tee: descarte por timeout, saida=2 valor=5
`)
	})
}
