package main

import (
	"fmt"
	"sync"
	"time"
)

// esperarTodas dispara n gorrotinas que fazem uma primeira fase, esperam
// as outras chegarem à barreira e só então seguem para a segunda. Cada
// uma é ao mesmo tempo esperada (Done) e quem espera (Wait).
func esperarTodas(n int) {
	var barreira sync.WaitGroup
	barreira.Add(n)

	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			// O Sleep é andaime: escalona as chegadas para a saída mostrar
			// que ninguém passa antes de a última chegar.
			time.Sleep(time.Duration(i) * 50 * time.Millisecond)
			fmt.Printf("gorrotina %d: chegou à barreira\n", i+1)

			barreira.Done()
			barreira.Wait() // libera quando a última chamar Done
			fmt.Printf("gorrotina %d: passou a barreira\n", i+1)
		})
	}
	wg.Wait()
}

func main() {
	esperarTodas(4)
}
