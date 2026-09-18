package main

import (
	"fmt"
	"math/rand/v2"
	"time"
)

// replica simula um servidor cuja latência varia a cada chamada.
func replica(nome string) func(string) string {
	return func(consulta string) string {
		time.Sleep(rand.N(100 * time.Millisecond))
		return fmt.Sprintf("%s respondeu a %q", nome, consulta)
	}
}

// primeiro envia a mesma consulta a todas as réplicas e devolve a primeira
// resposta que chegar.
func primeiro(consulta string, replicas ...func(string) string) string {
	// O buffer tem uma vaga por réplica: as respostas perdedoras são
	// depositadas sem bloquear. Sem ele, essas goroutines ficariam presas
	// no envio para sempre, pois ninguém mais vai ler do canal.
	c := make(chan string, len(replicas))
	for _, r := range replicas {
		go func() { c <- r(consulta) }()
	}
	return <-c
}

func main() {
	replicas := []func(string) string{
		replica("réplica 1"),
		replica("réplica 2"),
		replica("réplica 3"),
	}

	fmt.Println(primeiro("golang", replicas...))

	// Combinado com timeout: usa a resposta mais rápida, desde que chegue
	// em até 20ms. O buffer de tamanho 1 tem o mesmo papel: se o timeout
	// vencer, a goroutine ainda consegue depositar a resposta e terminar.
	resposta := make(chan string, 1)
	go func() { resposta <- primeiro("csp", replicas...) }()

	select {
	case r := <-resposta:
		fmt.Println(r)
	case <-time.After(20 * time.Millisecond):
		fmt.Println("tempo esgotado: nenhuma réplica respondeu em 20ms")
	}
}
