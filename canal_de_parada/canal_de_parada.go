package main

import "fmt"

// contador é um gerador sem fim: envia 0, 1, 2... até que o canal quit seja
// fechado. Ao sair, fecha o canal de saída.
func contador(quit <-chan struct{}) <-chan int {
	saida := make(chan int)
	go func() {
		defer close(saida)
		for i := 0; ; i++ {
			// O envio disputa com o sinal de parada: o que puder
			// prosseguir primeiro, vence.
			select {
			case saida <- i:
			case <-quit:
				return
			}
		}
	}()
	return saida
}

func main() {
	quit := make(chan struct{})
	valores := contador(quit)

	// Só queremos os três primeiros valores
	for range 3 {
		fmt.Println(<-valores)
	}

	// Manda o gerador parar. Sem isto ele ficaria bloqueado no próximo
	// envio para sempre.
	close(quit)

	// O gerador fecha a saída ao terminar: ler até o fechamento garante que
	// ele parou de fato.
	for range valores {
	}
	fmt.Println("o gerador parou")
}
