package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
)

// executarTarefasErrgroup faz o mesmo que executarTarefas, com um
// errgroup.Group no lugar do canal e do WaitGroup: SetLimit é o número de
// vagas, Go bloqueia quando elas acabam e Wait espera todas e devolve o
// primeiro erro. A tarefa de número `falha` devolve erro para mostrar o
// que acontece com as outras; ela é andaime, não faz parte do padrão.
func executarTarefasErrgroup(tarefas, limite, falha int) error {
	// O contexto é cancelado no primeiro erro
	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(limite)

	// Contador usado apenas para observar quantas tarefas estão ativas;
	// ele não faz parte do padrão.
	var ativas atomic.Int32

	for i := range tarefas {
		g.Go(func() error {
			// As tarefas que ainda não começaram desistem
			if ctx.Err() != nil {
				fmt.Printf("errgroup: tarefa %2d cancelada\n", i+1)
				return ctx.Err()
			}

			fmt.Printf("errgroup: tarefa %2d começou, ativas: %d\n", i+1, ativas.Add(1))
			defer ativas.Add(-1)

			// A falha vem antes do trabalho, de propósito. Se ela viesse
			// depois, as vizinhas terminariam no mesmo instante e poderiam
			// liberar vaga antes de o cancelamento chegar às próximas.
			if i+1 == falha {
				return fmt.Errorf("tarefa %d falhou", i+1)
			}
			time.Sleep(100 * time.Millisecond)
			return nil
		})
	}

	return g.Wait()
}
