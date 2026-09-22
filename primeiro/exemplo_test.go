package main

import (
	"context"
	"fmt"
	"testing"
	"testing/synctest"
	"time"

	"github.com/cassiobotaro/concorrencia-go/internal/saida"
)

// comLatencia cria uma réplica com latência fixa, para que o teste saiba
// de antemão quem responde primeiro.
func comLatencia(nome string, latencia time.Duration) func(context.Context, string) (string, error) {
	return func(ctx context.Context, consulta string) (string, error) {
		select {
		case <-time.After(latencia):
			return fmt.Sprintf("%s respondeu a %q", nome, consulta), nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}

// comErro cria uma réplica que falha depois de `latencia`, como um servidor
// fora do ar.
func comErro(nome string, latencia time.Duration) func(context.Context, string) (string, error) {
	return func(ctx context.Context, consulta string) (string, error) {
		select {
		case <-time.After(latencia):
			return "", fmt.Errorf("%s falhou", nome)
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}

func TestExemplo(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		obtida := saida.Capturar(t, func() {
			replicas := []func(context.Context, string) (string, error){
				comLatencia("réplica lenta", 300*time.Millisecond),
				comLatencia("réplica rápida", 10*time.Millisecond),
				comLatencia("réplica média", 150*time.Millisecond),
			}

			resposta, err := primeiro(context.Background(), "golang", replicas...)
			fmt.Println(resposta, err)

			// Com um prazo menor do que a réplica mais rápida, ninguém responde
			ctx, cancelar := context.WithTimeout(context.Background(), 5*time.Millisecond)
			defer cancelar()
			resposta, err = primeiro(ctx, "golang", replicas...)
			fmt.Printf("%q %v\n", resposta, err)

			// Se todas falharem, primeiro devolve o último erro em vez de
			// esperar para sempre por uma resposta que não vem
			resposta, err = primeiro(context.Background(), "golang",
				comErro("réplica 1", 0),
				comErro("réplica 2", 50*time.Millisecond),
			)
			fmt.Printf("%q %v\n", resposta, err)
		})
		saida.Conferir(t, obtida, `
réplica rápida respondeu a "golang" <nil>
"" context deadline exceeded
"" réplica 2 falhou
`)
	})
}
