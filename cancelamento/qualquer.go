package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
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
				// Outro sinal chegou primeiro: esta gorrotina termina
				// em vez de ficar presa esperando `c` para sempre.
			}
		}()
	}
	return saida
}

// trabalharAte imprime a cada 100ms até o canal parar ser fechado. Um
// único case de parada, não importa quantas origens existam.
func trabalharAte(parar <-chan struct{}) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			fmt.Println("trabalhando...")
		case <-parar:
			return
		}
	}
}

func combinarSinais() {
	// Três origens para o sinal de parada, e as três são contextos: um
	// deriva do outro, e cancelar o pai cancela os filhos.
	ctx, pararNoSinal := signal.NotifyContext(context.Background(), os.Interrupt)
	defer pararNoSinal()
	ctx, cancelarRequisicao := context.WithCancel(ctx) // um handler HTTP teria r.Context()
	defer cancelarRequisicao()
	ctx, cancelarPrazo := context.WithTimeout(ctx, 250*time.Millisecond)
	defer cancelarPrazo()

	trabalharAte(ctx.Done())
	fmt.Println("contexto cancelado:", ctx.Err())

	// qualquer fica para o que não é contexto. Aqui, um canal que outra
	// gorrotina fecha ao terminar; a gorrotina é andaime, simula um colega
	// que acaba antes do prazo.
	ctx, cancelar := context.WithTimeout(context.Background(), time.Second)
	defer cancelar()
	colegaTerminou := make(chan struct{})
	go func() {
		time.Sleep(150 * time.Millisecond)
		close(colegaTerminou)
	}()

	trabalharAte(qualquer(ctx.Done(), colegaTerminou))
	fmt.Println("um dos sinais de parada chegou (aqui, o colega terminou)")
}
