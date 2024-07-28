package main

import (
	"context"
	"fmt"
	"time"
)

// publicar tenta enviar um valor para o canal `saida` e utiliza um contexto com timeout
// para garantir que a operação não dure mais do que o tempo especificado.
func publicar(ctx context.Context, saida chan<- int, valor int, controle chan<- struct{}) {
	// Cria um contexto com timeout de 1 segundo
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		// Se o contexto expirar antes do envio, não faz nada
	case saida <- valor:
		// Se o valor for enviado com sucesso antes do timeout
	}
	controle <- struct{}{}
}

func fanout(entrada <-chan int, saidas ...chan<- int) {
	// Canal para controlar o término das publicações
	controle := make(chan struct{}, len(saidas)*2) // capacidade para controle de todas as publicações

	for valor := range entrada {
		// Publica o valor de entrada em todas as saídas
		for _, saida := range saidas {
			go publicar(context.Background(), saida, valor, controle)
		}
		// Aguarda o término de todas as publicações
		for i := 0; i < len(saidas); i++ {
			<-controle
		}
	}
	// Como a entrada foi consumida, fecha os canais de saída
	for _, saida := range saidas {
		close(saida)
	}
}

func sequenciaNumeros(inicial, final int) <-chan int {
	saida := make(chan int)
	go func() {
		for i := inicial; i <= final; i++ {
			saida <- i
		}
		// após gerar todos os valores, fecha o canal
		close(saida)
	}()
	return saida
}

func trabalhador(in <-chan int, id int, controle chan<- struct{}) {
	for v := range in {
		fmt.Println("id: ", id, " valor: ", v)
	}
	controle <- struct{}{}
}

func main() {
	saida1 := make(chan int)
	saida2 := make(chan int)

	// Canal para aguardar o término dos trabalhadores
	controle := make(chan struct{}, 2)

	// Inicia trabalhadores
	go trabalhador(saida1, 1, controle)
	go trabalhador(saida2, 2, controle)

	// Distribui a sequência de números para os canais de saída
	fanout(sequenciaNumeros(1, 10), saida1, saida2)

	// Aguarda o término dos trabalhadores
	for i := 0; i < 2; i++ {
		<-controle
	}
}
