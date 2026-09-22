package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// executarTarefas dispara uma gorrotina por tarefa, mas deixa no máximo
// `limite` delas executarem ao mesmo tempo.
func executarTarefas(tarefas, limite int) {
	// Um canal com buffer funciona como semáforo: cada valor no buffer é uma
	// vaga ocupada. Enviar bloqueia quando as `limite` vagas estão ocupadas.
	sem := make(chan struct{}, limite)

	var wg sync.WaitGroup
	// Contador usado apenas para observar quantas tarefas estão ativas;
	// ele não faz parte do padrão.
	var ativas atomic.Int32

	// Cada tarefa tem sua própria gorrotina, mas só `limite` avançam por vez
	for i := range tarefas {
		wg.Go(func() {
			sem <- struct{}{}        // ocupa uma vaga (bloqueia se não houver)
			defer func() { <-sem }() // libera a vaga ao terminar

			fmt.Printf("tarefa %2d começou, ativas: %d\n", i+1, ativas.Add(1))
			time.Sleep(100 * time.Millisecond)
			ativas.Add(-1)
		})
	}

	wg.Wait()
}

func main() {
	// Dez tarefas, no máximo três ao mesmo tempo
	executarTarefas(10, 3)

	// O mesmo com errgroup (veja com_errgroup.go): a tarefa 2 falha e as
	// que ainda não começaram são canceladas
	if err := executarTarefasErrgroup(10, 3, 2); err != nil {
		fmt.Println("errgroup:", err)
	}
}
