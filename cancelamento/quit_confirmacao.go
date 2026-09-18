package main

import (
	"fmt"
	"time"
)

// tagarelaComConfirmacao envia mensagens até receber algo no canal quit.
// Antes de sair faz a limpeza e confirma no MESMO canal que terminou,
// por isso o canal é bidirecional.
func tagarelaComConfirmacao(nome string, quit chan string) <-chan string {
	saida := make(chan string)
	go func() {
		for i := 0; ; i++ {
			select {
			case saida <- fmt.Sprintf("%s %d", nome, i):
			case <-quit:
				limpeza()
				quit <- "parei"
				return
			}
		}
	}()
	return saida
}

// limpeza simula a liberação de recursos: fechar arquivos, conexões etc.
func limpeza() {
	fmt.Println("gerador: liberando recursos...")
	time.Sleep(100 * time.Millisecond)
}

func quitComConfirmacao() {
	quit := make(chan string)
	c := tagarelaComConfirmacao("Duda", quit)
	for range 3 {
		fmt.Println(<-c)
	}
	quit <- "pare"
	// Só seguimos em frente depois que o gerador confirmar que terminou
	fmt.Println("gerador:", <-quit)
}
