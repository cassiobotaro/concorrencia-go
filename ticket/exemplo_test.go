package main

import (
	"context"
	"fmt"
	"testing"
	"testing/synctest"
	"time"

	"github.com/cassiobotaro/concorrencia-go/internal/saida"
)

// O main leva 3 segundos. Aqui a mesma montagem roda com 5 tickets a cada
// 50ms (um a cada 10ms): só o primeiro ticket sai de imediato, então três
// trabalhos levam dois intervalos. Na bolha do synctest o relógio é falso
// e só anda quando todas as gorrotinas esperam, por isso o tempo medido é
// exato, e não apenas um mínimo.
func TestExemplo(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		obtida := saida.Capturar(t, func() {
			tickets := make(chan ticket)
			trabalhos := make(chan func())
			pronto := make(chan struct{})

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go bilheteria(ctx, tickets, 50*time.Millisecond, 5)
			go func() {
				trabalhador(tickets, trabalhos)
				close(pronto)
			}()

			inicio := time.Now()
			for i := range 3 {
				trabalhos <- func() {
					fmt.Println("processando trabalho", i)
				}
			}
			close(trabalhos)
			<-pronto

			fmt.Println("tempo decorrido:", time.Since(inicio))
		})
		saida.Conferir(t, obtida, `
processando trabalho 0
processando trabalho 1
processando trabalho 2
tempo decorrido: 20ms
`)
	})
}
