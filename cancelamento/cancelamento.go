package main

import (
	"context"
	"fmt"
	"runtime"
)

// sequenciaNumeros é o gerador usado nos outros exemplos. Ele não é
// cancelável: se o consumidor parar de ler antes do fim, o envio bloqueia
// para sempre e a goroutine vaza.
func sequenciaNumeros(inicial, final int) <-chan int {
	saida := make(chan int)
	go func() {
		for i := inicial; i <= final; i++ {
			saida <- i
		}
		close(saida)
	}()
	return saida
}

// sequenciaNumerosCancelavel faz cada envio disputar com ctx.Done():
// se o contexto for cancelado, a goroutine desiste do envio e termina.
func sequenciaNumerosCancelavel(ctx context.Context, inicial, final int) <-chan int {
	saida := make(chan int)
	go func() {
		defer close(saida)
		for i := inicial; i <= final; i++ {
			select {
			case saida <- i:
			case <-ctx.Done():
				fmt.Println("gerador: cancelado, encerrando")
				return
			}
		}
	}()
	return saida
}

func main() {
	antes := runtime.NumGoroutine()

	// Sem cancelamento: lemos só os 3 primeiros valores e paramos.
	valores := sequenciaNumeros(1, 1000)
	for range 3 {
		fmt.Printf("valor: %v\n", <-valores)
	}
	// Ninguém mais vai ler de `valores`: a goroutine do gerador está presa
	// em `saida <- 4` e continuará assim até o programa terminar.
	fmt.Printf("goroutines presas: %d\n", runtime.NumGoroutine()-antes)

	// Com cancelamento: lemos os 3 primeiros valores e cancelamos.
	ctx, cancel := context.WithCancel(context.Background())
	cancelaveis := sequenciaNumerosCancelavel(ctx, 1, 1000)
	for range 3 {
		fmt.Printf("valor: %v\n", <-cancelaveis)
	}
	// Sem esta chamada a goroutine ficaria presa, como a anterior.
	cancel()
	// O gerador fecha o canal ao sair: drenar até o fechamento garante que
	// ele terminou de fato.
	for range cancelaveis {
	}
	fmt.Println("gerador cancelável encerrado")

	// Parada com confirmação (veja context_errgroup.go)
	paradaComErrgroup()
	// Vários sinais de parada combinados em um só (veja qualquer.go)
	combinarSinais()
}
