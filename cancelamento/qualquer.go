package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// qualquer combina vários sinais de parada em um só: o canal devolvido é
// fechado assim que o primeiro dos canais recebidos for fechado.
func qualquer(canais ...<-chan struct{}) <-chan struct{} {
	saida := make(chan struct{})
	// Mais de um sinal pode chegar ao mesmo tempo, e fechar um canal duas
	// vezes causa panic: o sync.Once garante um único close.
	var once sync.Once
	for _, c := range canais {
		go func() {
			select {
			case <-c:
				once.Do(func() { close(saida) })
			case <-saida:
				// Outro sinal chegou primeiro: esta goroutine termina
				// em vez de ficar presa esperando `c` para sempre.
			}
		}()
	}
	return saida
}

func combinarSinais() {
	// Três origens independentes para o sinal de parada
	ctxRequisicao, cancelarRequisicao := context.WithCancel(context.Background())
	defer cancelarRequisicao()
	ctxPrazo, cancelarPrazo := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancelarPrazo()
	desligar := make(chan struct{}) // seria fechado ao receber um sinal do sistema operacional

	parar := qualquer(ctxRequisicao.Done(), ctxPrazo.Done(), desligar)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		// Um único case de parada, não importa quantas origens existam
		select {
		case <-ticker.C:
			fmt.Println("trabalhando...")
		case <-parar:
			fmt.Println("um dos sinais de parada chegou (aqui, o prazo de 250ms)")
			return
		}
	}
}
