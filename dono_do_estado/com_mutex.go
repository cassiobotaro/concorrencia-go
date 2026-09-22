package main

import (
	"fmt"
	"sync"
)

// contadorMutex resolve o mesmo problema serializando o acesso ao mapa.
// Quando tudo o que se precisa é proteger um dado, esta versão é mais
// simples e mais clara do que uma gorrotina dona do estado.
type contadorMutex struct {
	mu       sync.Mutex
	contagem map[string]int
}

func (c *contadorMutex) incrementar(chave string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.contagem[chave]++
}

func (c *contadorMutex) consultar(chave string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.contagem[chave]
}

func comMutex() {
	c := contadorMutex{contagem: make(map[string]int)}

	var wg sync.WaitGroup
	for _, chave := range []string{"gopher", "gopher", "marmota"} {
		wg.Go(func() {
			for range 1000 {
				c.incrementar(chave)
			}
		})
	}
	wg.Wait()

	for _, chave := range []string{"gopher", "marmota"} {
		fmt.Printf("mutex: %s = %d\n", chave, c.consultar(chave))
	}
}
