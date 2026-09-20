package main

import (
	"fmt"
	"sync"
	"time"
)

// tarefa simula um trabalho que leva algum tempo.
func tarefa(id int) {
	time.Sleep(time.Duration(id) * 50 * time.Millisecond)
	fmt.Printf("tarefa %d terminou\n", id)
}

func main() {
	var wg sync.WaitGroup

	// wg.Go dispara a função em uma nova goroutine e registra no
	// WaitGroup que ela precisa terminar.
	for i := range 3 {
		wg.Go(func() {
			tarefa(i + 1)
		})
	}

	// Bloqueia até que todas as goroutines disparadas com wg.Go terminem.
	// Sem esta linha o programa acabaria antes de as tarefas imprimirem.
	wg.Wait()
	fmt.Println("todas as tarefas terminaram")
}
