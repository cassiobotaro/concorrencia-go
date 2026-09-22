package main

import (
	"os"
	"os/signal"
	"testing"
	"testing/synctest"

	"github.com/cassiobotaro/concorrencia-go/internal/saida"
)

func TestExemplo(t *testing.T) {
	// Ativar um sinal conversa com uma gorrotina do runtime que vive fora
	// da bolha, e dentro dela essa espera trava o teste. Registrando o
	// sinal aqui fora, o signal.NotifyContext de combinarSinais encontra
	// o sinal já ativo e não precisa fazer essa conversa.
	sinais := make(chan os.Signal, 1)
	signal.Notify(sinais, os.Interrupt)
	defer signal.Stop(sinais)

	synctest.Test(t, func(t *testing.T) {
		saida.Conferir(t, saida.Capturar(t, main), `
Bia 0
Bia 1
Bia 2
gerador: liberando recursos...
gerador: context canceled
trabalhando...
trabalhando...
contexto cancelado: context deadline exceeded
trabalhando...
um dos sinais de parada chegou (aqui, o colega terminou)
`)
	})
}
