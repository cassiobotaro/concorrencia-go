package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// limite é o número máximo de tarefas executando ao mesmo tempo.
const limite = 3

func main() {
	// Um canal com buffer funciona como semáforo: cada valor no buffer é uma
	// vaga ocupada. Enviar bloqueia quando as `limite` vagas estão ocupadas.
	sem := make(chan struct{}, limite)

	var wg sync.WaitGroup
	// Contador usado apenas para observar quantas tarefas estão ativas;
	// ele não faz parte do padrão.
	var ativas atomic.Int32

	// Cada tarefa tem sua própria goroutine, mas só `limite` avançam por vez
	wg.Add(10)
	for i := range 10 {
		go func() {
			defer wg.Done()

			sem <- struct{}{}        // ocupa uma vaga (bloqueia se não houver)
			defer func() { <-sem }() // libera a vaga ao terminar

			fmt.Printf("tarefa %2d começou, ativas: %d\n", i+1, ativas.Add(1))
			time.Sleep(100 * time.Millisecond)
			ativas.Add(-1)
		}()
	}

	wg.Wait()
}
