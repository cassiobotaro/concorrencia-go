# Concorrência em Go

Go é fundamentada no modelo CSP (Communicating sequential processes) proposto por Tony Hoare. Neste modelo os dados são compartilhados enviando mensagens através de canais.

As explicações e exemplos são altamente inspiradas na [apresentação](https://github.com/andrebq/andrebq.github.io) do @andrebq.

Uma outra influência é a [artigo](https://go.dev/blog/pipelines) sobre _pipelines_ e cancelmaneto em go.

## 🔗 Canais

Canais (channels) são uma estrutura primitiva na linguagem, e você pode utilizá-los para envio e recebimento de valores entre rotinas (_goroutines_). Os valores podem ser de qualquer tipo, inclusive do tipo canal.

Um canal é um ponto de sincronização entre _goroutines_. Uma _goroutine_ vai ficar bloqueada escrevendo em um canal até que aquele canal seja lido.

Ler de um canal é semelhante, uma _goroutine_ vai ficar bloqueada lendo até que um valor seja enviado para o canal ou o canal seja fechado (quando isso ocorre, o valor zero do tipo é retornado).

Um canal pode ser fechado. Isso é útil para indicar que nenhum outro valor seré escrito no canal.

Ler um canal fechado retorna um valor zero do tipo do canal.

Escrever em um canal fechado retorna um erro (_panic_).

## 🗺️ Olá Mundo

O [primeiro exemplo](./ola_mundo/ola_mundo.go) mostra como criar um canal, que será utilizado como ponte entre a aplicação principal e uma _goroutine_.

O programa principal fica bloqueado até que a mensagem "Olá mundo" seja enviada para o canal.

Quando isto ocorre, a _goroutine_ é desbloqueada e a mensagem é exibida.

Quando o programa principal termina, a _goroutine_ é também terminada.

```go
package main

import "fmt"

func main() {
	canal := make(chan string)
	go func() {
		canal <- "Olá, mundo!"
	}()

	fmt.Println(<-canal)
}
```

## 🆕 Geradores

Geradores são funções que iniciam uma _goroutine_ para escrever uma lista de valores em um canal que é retornado para quem acionou a função.

No exemplo uma sequência de números inteiros é gerada e enviada para um canal.

A função principal (_main_) irá realizar a leitura do canal e imprimir os valores. Essa é uma característica interessante sobre canais, quando utilizados com o _range_, a iteração continuará até que o canal seja fechado.

```go
package main

import "fmt"

func sequencia_numeros(inicial, final int) <-chan int {
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

func main() {
	valores := sequencia_numeros(1, 1000)
	for valor := range valores {
		fmt.Printf("valor: %v\n", valor)
	}
}

```

## 👷 Pipeline

Um _pipeline_ trabalha recebendo valores de um canal e escrevendo em outro canal, normalmente após realizar alguma tranformação no valor.

No exemplo temos a função `dobroFloat` atuando como um _pipeline_, que irá receber os valores enviados ao canal de entrada retornando os valores transformados.

Um canal pode ser definido como sendo apenas para leitura (`<-`) ou apenas para escrita (`<-`).

Os valores gerados pelo gerador `sequencia_numeros` são enviados para o canal de entrada do pipeline e seu valor transformado recebido pelo canal de saída na função principal e é impresso.

Vários pipelines poderiam ser encadeados para realizar múltiplas transformações.

```go
package main

import "fmt"

func dobroFloat(entrada <-chan int) chan<- float64 {
	saida := make(chan float64)
	go func() {
		for valor := range entrada {
			saida <- float64(valor) * 2
		}
		// Após ter terminado de transformar os valores de entrada,
		//  fecha o canal de saida
		close(saida)
	}()
	return saida
}

func sequencia_numeros(inicial, final int) <-chan int {
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

func main() {
	for valor := range sequencia_numeros(1, 10) {
		fmt.Printf("valor: %v\n", valor)
	}
}

```

## ⚗️ Fan-in

Um fan-in copia dados de múltiplos canais de entrada e escreve em um único canal de saída. Normalmente um fan-in só termina quando todos os canais de entrada são fechados.

A função fan-in pode receber vários canais entrada através de [parâmetros múltiplos](https://gobyexample.com/variadic-functions).

No exemplo abaixo, enviamos vários geradores como entrada para a função fan-in e nos é retornado um único canal de saída. Internamente, uma _goroutine_ é criada para ler os valores de cada canal de entrada, porém todas escrevem no mesmo canal de saída.

Envio de mensagem em um canal fechado causa um erro (_panic_), por isso é importante garantir que todos os canais de entrada estejam fechados antes de fechar o canal de saída. O tipo sync.WaitGroup fornece uma maneira simples de organizar essa sincronização.

Repare que temos uma _goroutine_ que aguarda um sinal indicando que todas as todas entradas foram consumidas (wg.Wait), finalizando assim o canal de saída.

```go
package main

import (
	"fmt"
	"sync"
)

func fanin(canais_entrada ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	// canal de saída que será compartilhado entre os canais de entrada
	canal_saida := make(chan int)

	// lê os valores de cada canal de entrada e envia para o canal de saída
	// quando todos os valores forem lidos, envia sinal avisando que terminou
	output := func(c <-chan int) {
		for n := range c {
			canal_saida <- n
		}
		// aviso que terminou de ler os valores de um canal
		wg.Done()
	}
	wg.Add(len(canais_entrada))
	// Inicializa uma goroutine de saída para cada canal de entrada em canais_entrada.
	for _, c := range canais_entrada {
		go output(c)
	}

	// Inicia uma goroutine para fechar o canal de saída quando todas as
	// goroutines de entrada terminarem.
	// isto deve ser feito após o wg.Add
	go func() {
		wg.Wait()
		close(canal_saida)
	}()
	return canal_saida
}

func sequencia_numeros(inicial, final int) <-chan int {
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

func main() {
	canal := fanin(
		sequencia_numeros(1, 10),
		sequencia_numeros(11, 20),
		sequencia_numeros(21, 30),
	)
	for valor := range canal {
		fmt.Printf("valor: %v\n", valor)
	}
}

```

## 📣 Fan-out

Em breve

## 🪟 Janela deslizante

Em breve

## 🧑‍🤝‍🧑 Operções em volume/lote

Em breve

## 🎫 Sistema de ticket

Em breve

## 👨‍💻 Run On My goroutine

Em breve

## 🥸 Modelo de atores

Em breve

## 📰 Contextos

Em breve