package main

import (
	"fmt"
	"time"
)

// tagarela é um gerador que fala cada vez mais devagar: a pausa entre as
// mensagens cresce 100ms a cada envio. Ele só para quando o canal quit é
// fechado (veja o exemplo canal_de_parada).
func tagarela(nome string, quit <-chan struct{}) <-chan string {
	saida := make(chan string)
	go func() {
		defer close(saida)
		for i := 0; ; i++ {
			// O envio disputa com o sinal de parada: o que puder
			// prosseguir primeiro, vence.
			select {
			case saida <- fmt.Sprintf("%s %d", nome, i):
				time.Sleep(time.Duration(i) * 100 * time.Millisecond)
			case <-quit:
				return
			}
		}
	}()
	return saida
}

// timeoutPorMensagem desiste quando UMA mensagem demora mais do que 350ms.
// O time.After é criado a cada volta do laço, então o prazo é renovado
// sempre que uma mensagem chega.
func timeoutPorMensagem() {
	quit := make(chan struct{})
	defer close(quit)
	c := tagarela("Ana", quit)

	for {
		select {
		case s := <-c:
			fmt.Println(s)
		case <-time.After(350 * time.Millisecond):
			fmt.Println("Ana demorou demais para falar.")
			return
		}
	}
}

// timeoutDaConversa limita a duração da conversa INTEIRA a 500ms.
// O time.After é criado uma única vez, fora do laço: o prazo não é renovado.
func timeoutDaConversa() {
	quit := make(chan struct{})
	defer close(quit)
	c := tagarela("Beto", quit)

	timeout := time.After(500 * time.Millisecond)
	for {
		select {
		case s := <-c:
			fmt.Println(s)
		case <-timeout:
			fmt.Println("A conversa com o Beto acabou.")
			return
		}
	}
}

func main() {
	timeoutPorMensagem()
	timeoutDaConversa()
}
