package main

import "context"

// faninSelect combina um número fixo de canais de entrada (aqui, dois) usando
// uma única goroutine e um select, em vez de uma goroutine por entrada.
// Como só uma goroutine escreve na saída, ela mesma fecha o canal ao terminar:
// não é preciso contar ninguém.
func faninSelect(ctx context.Context, entrada1, entrada2 <-chan int) <-chan int {
	saida := make(chan int)
	go func() {
		defer close(saida)
		for entrada1 != nil || entrada2 != nil {
			var valor int
			var ok bool
			select {
			case valor, ok = <-entrada1:
				if !ok {
					// Entrada fechada: um canal nil nunca é selecionado,
					// o que desabilita este case.
					entrada1 = nil
					continue
				}
			case valor, ok = <-entrada2:
				if !ok {
					entrada2 = nil
					continue
				}
			case <-ctx.Done():
				return
			}
			select {
			case saida <- valor:
			case <-ctx.Done():
				return
			}
		}
	}()
	return saida
}
