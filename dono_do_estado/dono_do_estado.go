package main

import (
	"fmt"
	"sync"
)

// consulta pede a contagem de uma chave e carrega o canal de resposta,
// como no exemplo de requisição e resposta.
type consulta struct {
	chave    string
	resposta chan<- int
}

// contador é a gorrotina dona do estado: só ela toca o mapa `contagem`.
// As demais gorrotinas pedem alterações e leituras pelos canais.
// Termina quando o canal `incrementar` é fechado.
func contador(incrementar <-chan string, consultar <-chan consulta) {
	contagem := make(map[string]int)
	for {
		select {
		case chave, ok := <-incrementar:
			if !ok {
				return
			}
			contagem[chave]++
		case c := <-consultar:
			c.resposta <- contagem[c.chave]
		}
	}
}

func main() {
	incrementar := make(chan string)
	consultar := make(chan consulta)
	pronto := make(chan struct{})
	go func() {
		contador(incrementar, consultar)
		close(pronto)
	}()

	// Várias gorrotinas incrementam ao mesmo tempo, sem mutex:
	// os pedidos são atendidos um por vez pela gorrotina dona.
	var wg sync.WaitGroup
	for _, chave := range []string{"gopher", "gopher", "marmota"} {
		wg.Go(func() {
			for range 1000 {
				incrementar <- chave
			}
		})
	}
	wg.Wait()

	for _, chave := range []string{"gopher", "marmota"} {
		resposta := make(chan int)
		consultar <- consulta{chave: chave, resposta: resposta}
		fmt.Printf("dona do estado: %s = %d\n", chave, <-resposta)
	}

	// Sem mais incrementos: a gorrotina dona termina
	close(incrementar)
	<-pronto

	// A mesma contagem feita com mutex (veja com_mutex.go)
	comMutex()
}
