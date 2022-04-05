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

func main() {
	valores := sequenciaNumeros(1, 1000)
	for valor := range valores {
		fmt.Printf("valor: %v\n", valor)
	}
}

```

## 🚧 Consumidores (workers)

Um consumidor é uma _goroutine_ que recebe valores de um canal e os processa.

No exemplo valores inteiros são enviados pela função principal (main) através do canal de entrada e processados por um trabalhador.

É possível criar vários trabalhadores para processarem um mesmo canal.

```go
package main

import "fmt"

func trabalhador(canal_entrada <-chan int) {
	for valor := range canal_entrada {
		fmt.Printf("valor: %v\n", valor)
	}
}

func main() {
	entrada := make(chan int)
	// Um trabalhador é iniciado e aguarda por valores no canal de entrada
	go trabalhador(entrada)
	for i := 0; i < 10; i++ {
		entrada <- i
	}
	// Após ter enviado todos os valores, fecha o canal de entrada
	// avisando ao trabalhador que o trabalho terminou
	close(entrada)
}


```

## 👷 Pipeline

Um _pipeline_ trabalha recebendo valores de um canal e escrevendo em outro canal, normalmente após realizar alguma tranformação no valor.

No exemplo temos a função `dobroFloat` atuando como um _pipeline_, que irá receber os valores enviados ao canal de entrada retornando os valores transformados.

Um canal pode ser definido como sendo apenas para leitura (`<-`) ou apenas para escrita (`<-`).

Os valores gerados pelo gerador `sequenciaNumeros` são enviados para o canal de entrada do pipeline e seu valor transformado recebido pelo canal de saída na função principal e é impresso.

Vários pipelines poderiam ser encadeados para realizar múltiplas transformações.

```go
package main

import "fmt"

func dobroFloat(entrada <-chan int) <-chan float64 {
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

func main() {
	for valor := range dobroFloat(sequenciaNumeros(1, 10)) {
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

func fanin(entradas ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	// canal de saída que será compartilhado entre os canais de entrada
	saida := make(chan int)

	// lê os valores de cada canal de entrada e envia para o canal de saída
	// quando todos os valores forem lidos, envia sinal avisando que terminou
	enviarSaida := func(c <-chan int) {
		for n := range c {
			saida <- n
		}
		// aviso que terminou de ler os valores de um canal
		wg.Done()
	}
	wg.Add(len(entradas))
	// Inicializa uma goroutine de saída para cada canal de entrada em canais_entrada.
	for _, c := range entradas {
		go enviarSaida(c)
	}

	// Inicia uma goroutine para fechar o canal de saída quando todas as
	// goroutines de entrada terminarem.
	// isto deve ser feito após o wg.Add
	go func() {
		wg.Wait()
		close(saida)
	}()
	return saida
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

func main() {
	canal := fanin(
		sequenciaNumeros(1, 10),
		sequenciaNumeros(11, 20),
		sequenciaNumeros(21, 30),
	)
	for valor := range canal {
		fmt.Printf("valor: %v\n", valor)
	}
}

```

## 📣 Fan-out

Um fan-out copia dados de um canal de entrada para múltiplos canais de saída.

No exemplo uma sequência de números é gerada e enviada para múltiplos canais de saída. Estes canais possuem seus respectivos trabalhadores que irão fazer o processamento do valor.

Esta implementação de fan-out tenta garantir a entrega de todas as mensagens utilizando um agrupador (WaitGroup) para aguardar a publicação dos valores em todos os canais de saída. A publicação é feita em sua própria _goroutine_ e conta também com um mecanismo(_timer_) de forma a previnir o bloqueio caso algum canal de saída não consiga consumir a mensagem. As mensagens não consumidas são descartadas.

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func publicar(saida chan<- int, valor int, wg *sync.WaitGroup) {
	timer := time.NewTimer(1 * time.Second)
	// Aguarda 1 segundo ou o canal ser lido
	select {
	case saida <- valor:
	case <-timer.C:
	}
	// Independente do canal ser lido ou não,
	// avisa que a publicação terminou
	wg.Done()
	timer.Stop()
}

func fanout(entrada <-chan int, saidas ...chan<- int) {

	// O agrupamento das publicações é para evitar que
	// o processamento fique bloquando enquanto um canal de saída não é lido
	// e garante que todos os valores serão publicados
	var wg sync.WaitGroup
	for valor := range entrada {
		wg.Add(len(saidas))
		// Publica o valor de entrada em todas as saídas
		for _, saida := range saidas {
			go publicar(saida, valor, &wg)
		}
		wg.Wait()
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

func trabalhador(in <-chan int, id int, wg *sync.WaitGroup) {
	for v := range in {
		fmt.Println("id: ", id, " valor: ", v)
	}
	wg.Done()
}

func main() {
	saida1 := make(chan int)
	saida2 := make(chan int)
	// Agrupamos os trabalhadores de forma
	// a aguardar o processamento de todos antes do programa principal
	// ser finalizado
	var wg sync.WaitGroup
	wg.Add(2)
	go trabalhador(saida1, 1, &wg)
	go trabalhador(saida2, 2, &wg)
	fanout(sequenciaNumeros(1, 10), saida1, saida2)
	wg.Wait()
}

```

## 🪟 Janela deslizante

Uma janela deslizante (sliding window) é utilizada para prevenir que um leitor lento trave um escritor rápido. Ela funciona deslizando sobre os dados. A ordem de entregas é garantida porém dados antigos podem ser descartados se o consumidor for muito lento.

No exemplo uma sequência de números é gerada, porém nosso consumidor é mais lento que o produtor, logo a medida que a janela desliza os valores antigos são descartados.

Para fazer a janela deslizante, utilizamos um buffer, que possui um tamanho fixo. Utilizamos uma técnica de seleção (select) onde caso o canal de saída seja lido, enviamos o valor para o consumidor e o removemos do buffer. Caso o canal de entrada seja lido, o valor é adicionado ao buffer.


```go
package main

import (
	"container/list"
	"fmt"
	"time"
)

func janelaDeslizante(saida chan<- interface{}, entrada <-chan interface{}, tamanho int) {
	buffer := list.New()
	defer close(saida)
	for entrada != nil || buffer.Len() > 0 {
		if buffer.Len() == 0 {
			// nós temos um buffer vazio
			// e um canal de entrada válido
			val := <-entrada
			if val == nil { // assume que nil significa fechado
				entrada = nil // não vai mais ler dados
				continue
			}
			buffer.PushBack(val)
			continue
		}
		select {
		case saida <- buffer.Front().Value:
			// consumidor lê o dado
			buffer.Remove(buffer.Front()) // remove first item
		case val := <-entrada:
			// recebeu nova entrada
			if val == nil {
				// invalida entrada
				entrada = nil
				// continua já que podemos ter dados
				// no buffer
				continue
			}
			if buffer.Len() == tamanho {
				// buffer cheio, descarta dados antigos
				buffer.Remove(buffer.Front())
			}
			// adiciona novo dado no buffer
			buffer.PushBack(val)
		}
	}
}

func leitorLento(in <-chan interface{}) {
	for val := range in {
		fmt.Printf("valor: %v\n", val)
		time.Sleep(4 * time.Second)
	}
}

func sequenciaNumeros(inicial, final int) <-chan interface{} {
	saida := make(chan interface{})
	go func() {
		for i := inicial; i <= final; i++ {
			saida <- i
			time.Sleep(1 * time.Second)
		}
		// após gerar todos os valores, fecha o canal
		close(saida)
	}()
	return saida
}

func main() {
	valores := sequenciaNumeros(1, 10)
	saida := make(chan interface{})
	go leitorLento(saida)
	janelaDeslizante(saida, valores, 3)
}

```

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