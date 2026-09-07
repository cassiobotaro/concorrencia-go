package main

import (
	"fmt"
	"time"
)

// produtor gera valores o mais rápido que consegue. Ele não sabe nada sobre a
// velocidade do consumidor: quem o freia é a capacidade limitada do canal.
// Quando o buffer enche, o envio bloqueia e o produtor passa a andar no ritmo
// do consumidor. Essa é a contrapressão (backpressure).
func produtor(saida chan<- int, n int) {
	defer close(saida)
	for i := 1; i <= n; i++ {
		inicio := time.Now()
		saida <- i
		if espera := time.Since(inicio); espera > 10*time.Millisecond {
			fmt.Printf("Produtor: buffer cheio, esperou %v para enviar %d\n", espera.Round(time.Millisecond), i)
			continue
		}
		fmt.Printf("Produtor: enviou %d\n", i)
	}
}

// consumidorLento simula um trabalho que leva mais tempo do que a produção,
// como escrever em disco ou chamar um serviço externo.
func consumidorLento(entrada <-chan int, pronto chan<- struct{}) {
	// Fechar o canal é o idioma para sinalizar um evento único
	defer close(pronto)
	for valor := range entrada {
		fmt.Printf("Consumidor: processando %d\n", valor)
		time.Sleep(200 * time.Millisecond)
	}
}

func main() {
	// A capacidade do canal é a folga permitida entre produtor e consumidor:
	// até 3 valores podem esperar na fila. Além disso, o produtor bloqueia.
	// Um buffer sem limite deixaria o produtor correr à frente e a memória
	// crescer sem controle; aqui a fila tem um teto conhecido.
	fila := make(chan int, 3)
	pronto := make(chan struct{})

	inicio := time.Now()
	go consumidorLento(fila, pronto)
	produtor(fila, 10)
	<-pronto

	fmt.Printf("Fim da execução em %v: nada foi descartado, o produtor apenas esperou.\n", time.Since(inicio).Round(100*time.Millisecond))
}
