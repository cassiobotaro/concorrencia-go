package main

import (
	"context"
	"fmt"
	"time"
)

// trabalhador produz um resultado a cada 3 intervalos e, enquanto isso,
// emite um batimento a cada intervalo para mostrar que continua vivo.
// No terceiro resultado ele trava por `travamento`, e os batimentos param.
func trabalhador(ctx context.Context, intervalo, travamento time.Duration) (<-chan struct{}, <-chan int) {
	batimento := make(chan struct{})
	resultados := make(chan int)
	go func() {
		defer close(resultados)
		pulso := time.NewTicker(intervalo)
		defer pulso.Stop()
		trabalho := time.NewTicker(3 * intervalo)
		defer trabalho.Stop()

		for i := 1; ; {
			select {
			case <-ctx.Done():
				return
			case <-pulso.C:
				// Envio não bloqueante: se ninguém estiver ouvindo,
				// o batimento é perdido e o trabalho segue.
				select {
				case batimento <- struct{}{}:
				default:
				}
			case <-trabalho.C:
				if i == 3 {
					// Simula um travamento
					time.Sleep(travamento)
				}
				select {
				case resultados <- i:
					i++
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return batimento, resultados
}

// supervisor acompanha o trabalhador e o declara morto se ficar dois
// intervalos sem batimento e sem resultado.
func supervisor(intervalo, travamento time.Duration) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	batimento, resultados := trabalhador(ctx, intervalo, travamento)

	for {
		select {
		case <-batimento:
			fmt.Println("batimento")
		case r := <-resultados:
			fmt.Println("resultado:", r)
		case <-time.After(2 * intervalo):
			fmt.Println("trabalhador não responde")
			return
		}
	}
}

func main() {
	supervisor(100*time.Millisecond, 1*time.Second)
}
