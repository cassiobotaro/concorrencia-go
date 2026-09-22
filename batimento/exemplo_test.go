package main

import (
	"context"
	"fmt"
	"time"
)

// O main imprime um "batimento" por sinal recebido, mas o envio é não
// bloqueante: um batimento que chega enquanto o supervisor está imprimindo
// se perde, e contar exatamente quantos chegaram dependeria do escalonador.
// Aqui o supervisor é refeito para contar em vez de imprimir, e a saída diz
// só o que é garantido: houve batimentos, vieram dois resultados e depois
// o silêncio.
func Example() {
	intervalo := 100 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	batimento, resultados := trabalhador(ctx, intervalo, time.Second)

	batimentos := 0
	for vivo := true; vivo; {
		select {
		case <-batimento:
			batimentos++
		case r := <-resultados:
			fmt.Println("resultado:", r)
		case <-time.After(2 * intervalo):
			fmt.Println("trabalhador não responde")
			vivo = false
		}
	}
	fmt.Println("houve batimentos antes de travar:", batimentos > 0)

	// Output:
	// resultado: 1
	// resultado: 2
	// trabalhador não responde
	// houve batimentos antes de travar: true
}
