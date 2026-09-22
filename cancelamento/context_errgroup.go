package main

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

// tagarelaComContext envia mensagens até o contexto ser cancelado. Antes
// de sair faz a limpeza e devolve o motivo do cancelamento. A confirmação
// é o próprio retorno: quem chamou espera por ele com g.Wait.
func tagarelaComContext(ctx context.Context, nome string, saida chan<- string) error {
	for i := 0; ; i++ {
		select {
		case saida <- fmt.Sprintf("%s %d", nome, i):
		case <-ctx.Done():
			limpeza()
			return ctx.Err()
		}
	}
}

func paradaComErrgroup() {
	ctx, cancelar := context.WithCancel(context.Background())
	saida := make(chan string)

	var g errgroup.Group
	g.Go(func() error { return tagarelaComContext(ctx, "Bia", saida) })

	for range 3 {
		fmt.Println(<-saida)
	}
	// cancelar manda parar. Wait bloqueia até a goroutine retornar e traz
	// o erro dela: o pedido desce pelo contexto e a resposta sobe pelo
	// errgroup.
	cancelar()
	fmt.Println("gerador:", g.Wait())
}
