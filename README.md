# Concorrência em Go

Go é fundamentada no modelo CSP (Communicating sequential processes) proposto por Tony Hoare. Neste modelo, os dados são compartilhados enviando mensagens através de canais.

As explicações e exemplos são altamente inspirados na [apresentação](https://github.com/andrebq/andrebq.github.io) do @andrebq.

Uma outra influência é o [artigo](https://go.dev/blog/pipelines) sobre _pipelines_ e cancelamento em Go.

Aqui serão apresentados alguns padrões de concorrência, porém sugiro também a leitura sobre [context](https://gobyexample.com/context), [select](https://gobyexample.com/select), [canais com buffer](https://gobyexample.com/channel-buffering) e outros mecanismos de controle de concorrência.

## 🔗 Canais

Canais (channels) são uma estrutura primitiva na linguagem, e você pode utilizá-los para envio e recebimento de valores entre rotinas (_goroutines_). Os valores podem ser de qualquer tipo, inclusive do tipo canal.

Um canal é um ponto de sincronização entre _goroutines_. Uma _goroutine_ vai ficar bloqueada escrevendo em um canal até que aquele canal seja lido.

Ler de um canal é semelhante, uma _goroutine_ vai ficar bloqueada lendo até que um valor seja enviado para o canal ou o canal seja fechado (quando isso ocorre, o valor zero do tipo é retornado).

Um canal pode ser fechado. Isso é útil para indicar que nenhum outro valor será escrito no canal.

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

No exemplo, uma sequência de números inteiros é gerada e enviada para um canal.

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

## 🚧 Trabalhador (worker)

Um trabalhador é uma _goroutine_ que recebe valores de um canal e os processa.

No exemplo, valores inteiros são enviados pela função principal (main) através do canal de entrada e processados por um trabalhador.

É possível criar vários trabalhadores para processarem um mesmo canal.

```go
package main

import "fmt"

func trabalhador(entrada <-chan int) {
	for valor := range entrada {
		fmt.Printf("valor: %v\n", valor)
	}
}

func main() {
	entrada := make(chan int)
	pronto := make(chan struct{})
	// Um trabalhador é iniciado e aguarda por valores no canal de entrada
	go func() {
		trabalhador(entrada)
		pronto <- struct{}{}
	}()
	for i := range 10 {
		entrada <- i
	}
	// Após ter enviado todos os valores, fecha o canal de entrada
	// avisando ao trabalhador que o trabalho terminou
	close(entrada)
	// Aguarda o trabalhador terminar
	<-pronto
}
```

## 👷‍♂️👷‍♀️ Grupo de Trabalhadores (pool of workers)

A piscina de marmotinhas (carinhosamente chamada pela minha esposa) é uma coleção de _goroutines_ que ficam esperando tarefas serem atribuídas a elas. Quando a _goroutine_ finaliza a tarefa que foi atribuída, se torna disponível novamente para execução de uma nova tarefa.

No exemplo, um grupo de n trabalhadores aguarda a chegada de valores pelo canal de entrada. Cada trabalhador executa seu processamento e envia o resultado por um canal.

Um canal de sinalização é utilizado para indicar que todos os trabalhadores terminaram.

```go
package main

import (
	"fmt"
)

// trabalhador processa valores recebidos do canal de entrada e envia resultados para o canal de saída.
// Ele utiliza um canal de sinalização para notificar quando terminar.
func trabalhador(id int, entrada <-chan int, saida chan<- int, terminar chan struct{}) {
	for valor := range entrada {
		fmt.Printf("id: %d processou valor: %v\n", id, valor)
		saida <- valor * 2
	}

	// Envia uma mensagem para o canal de sinalização ao terminar
	fmt.Printf("id: %d terminou\n", id)
	terminar <- struct{}{}
}

func grupoDeTrabalhadores(entrada <-chan int, nTrabalhadores int) chan int {
	saida := make(chan int)
	terminar := make(chan struct{}, nTrabalhadores)

	// Cria e inicia os trabalhadores
	for i := range nTrabalhadores {
		go trabalhador(i+1, entrada, saida, terminar)
	}

	// Goroutine para fechar o canal de saída quando todos os trabalhadores terminarem
	go func() {
		// Espera receber sinais de todos os trabalhadores
		for range nTrabalhadores {
			<-terminar
		}
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
		// Após gerar todos os valores, fecha o canal
		close(saida)
	}()
	return saida
}

func main() {
	// Produz uma sequência de 10 valores
	entrada := sequenciaNumeros(1, 10)
	// Um grupo de trabalhadores irá processar esses números
	saida := grupoDeTrabalhadores(entrada, 2)

	// Somente termina quando todo o trabalho for processado
	for s := range saida {
		fmt.Println(s)
	}
}
```

## 🧑‍🏭 Pipeline

Um _pipeline_ trabalha recebendo valores de um canal e escrevendo em outro canal, normalmente após realizar alguma transformação no valor.

No exemplo temos a função `dobro` atuando como um _pipeline_, que irá receber os valores enviados ao canal de entrada retornando os valores transformados.

Um canal pode ser definido como sendo apenas para leitura (`<-chan`) ou apenas para escrita (`chan<-`).

Os valores gerados pelo gerador `sequenciaNumeros` são enviados para o canal de entrada do pipeline e seu valor transformado recebido pelo canal de saída na função principal e é impresso.

Vários pipelines poderiam ser encadeados para realizar múltiplas transformações.

```go
package main

import "fmt"

func dobro(entrada <-chan int) <-chan int {
	saida := make(chan int)
	go func() {
		for valor := range entrada {
			saida <- valor * 2
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
	for valor := range dobro(dobro(sequenciaNumeros(1, 10))) {
		fmt.Printf("valor: %v\n", valor)
	}
}
```

## ⚗️ Fan-in

Um fan-in copia dados de múltiplos canais de entrada e escreve em um único canal de saída. Normalmente um fan-in só termina quando todos os canais de entrada são fechados.

A função fan-in pode receber vários canais de entrada através de [parâmetros múltiplos](https://gobyexample.com/variadic-functions).

No exemplo abaixo, enviamos vários geradores como entrada para a função fan-in e nos é retornado um único canal de saída. Internamente, uma _goroutine_ é criada para ler os valores de cada canal de entrada, porém todas escrevem no mesmo canal de saída.

Envio de mensagem em um canal fechado causa um erro (_panic_), por isso é importante garantir que todos os canais de entrada estejam fechados antes de fechar o canal de saída. Utilizamos um canal de sinalização para indicar que todos os canais de entrada foram processados.

Repare que temos uma _goroutine_ que aguarda um sinal indicando que todas as entradas foram consumidas, finalizando assim o canal de saída.

```go
package main

import (
	"fmt"
)

// fanin combina vários canais de entrada em um único canal de saída.
// Utiliza um canal de sinalização para saber quando todos os canais de entrada foram processados.
func fanin(entradas ...<-chan int) <-chan int {
	saida := make(chan int)
	// Número de canais de entrada
	n := len(entradas)
	// Canal de controle para quando todos os canais de entrada terminarem
	canalTermino := make(chan struct{}, n)

	for _, c := range entradas {
		go func(c <-chan int) {
			for n := range c {
				saida <- n
			}
			// Notifica que este canal foi processado
			canalTermino <- struct{}{}
		}(c)
	}

	// Quando todos os canais de entrada terminarem, fecha o canal de saída
	go func() {
		for range n {
			<-canalTermino
		}
		close(saida)
	}()

	return saida
}

// sequenciaNumeros cria um canal que envia uma sequência de números de inicial a final.
func sequenciaNumeros(inicial, final int) <-chan int {
	saida := make(chan int)
	go func() {
		for i := inicial; i <= final; i++ {
			saida <- i
		}
		close(saida)
	}()
	return saida
}

func main() {
	// Combina três canais de sequência em um único canal
	canal := fanin(
		sequenciaNumeros(1, 10),
		sequenciaNumeros(11, 20),
		sequenciaNumeros(21, 30),
	)

	// Lê e imprime os valores do canal combinado
	for valor := range canal {
		fmt.Printf("valor: %v\n", valor)
	}
}
```

## 📣 Fan-out

Um fan-out copia dados de um canal de entrada para múltiplos canais de saída.

No exemplo, uma sequência de números é gerada e enviada para múltiplos canais de saída. Estes canais possuem seus respectivos trabalhadores que irão fazer o processamento do valor.

Esta implementação de fan-out tenta garantir a entrega de todas as mensagens utilizando um agrupador (WaitGroup) para aguardar a publicação dos valores em todos os canais de saída. A publicação é feita em sua própria _goroutine_ e conta também com um mecanismo (_timer_) de forma a prevenir o bloqueio caso algum canal de saída não consiga consumir a mensagem. As mensagens não consumidas são descartadas.

```go
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
		for range saidas {
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
	for range 2 {
		<-controle
	}
}
```

## 🪟 Janela deslizante

Uma janela deslizante (sliding window) é utilizada para prevenir que um leitor lento trave um escritor rápido. Ela funciona deslizando sobre os dados. A ordem de entregas é garantida, porém dados antigos podem ser descartados se o consumidor for muito lento.

No exemplo, uma sequência de números é gerada, porém nosso consumidor é mais lento que o produtor, logo à medida que a janela desliza os valores antigos são descartados.

Para fazer a janela deslizante, utilizamos um buffer, que possui um tamanho fixo. Utilizamos uma técnica de seleção (select) onde, caso o canal de saída seja lido, enviamos o valor para o consumidor e o removemos do buffer. Caso o canal de entrada seja lido, o valor é adicionado ao buffer.

```go
package main

import (
	"fmt"
	"time"
)

func janelaDeslizante(saida chan<- any, entrada <-chan any, tamanho int) {
	buffer := make(chan any, tamanho)
	defer close(saida)

	// Lógica de leitura do produtor
	go func() {
		defer close(buffer)
		for val := range entrada {
			// Tenta enviar para o buffer
			select {
			case buffer <- val:
				// Enviou com sucesso
			default:
				// Buffer cheio, descarta o mais antigo e adiciona o novo
				<-buffer
				fmt.Printf("Janela Deslizante: Buffer cheio, descartou valor antigo para adicionar %v.\n", val)
				buffer <- val
			}
		}
	}()

	// Lógica de envio para o consumidor
	for val := range buffer {
		saida <- val
		fmt.Printf("Janela Deslizante: Enviou %v para o consumidor.\n", val)
	}
}

// O resto do código permanece o mesmo.
func sequenciaNumeros(inicial, final int) <-chan any {
	saida := make(chan any)
	go func() {
		for i := inicial; i <= final; i++ {
			saida <- i
			fmt.Printf("Produtor: Enviou %d\n", i)
			time.Sleep(1 * time.Second)
		}
		close(saida)
	}()
	return saida
}

func leitorLento(in <-chan any) {
	for val := range in {
		fmt.Printf("Consumidor: Recebeu %v\n", val)
		time.Sleep(4 * time.Second)
	}
}

func main() {
	valores := sequenciaNumeros(1, 10)
	saida := make(chan any)
	go leitorLento(saida)
	janelaDeslizante(saida, valores, 3)
	fmt.Println("Fim da execução.")
}
```

## 🧑‍🤝‍🧑 Processamento em lote (batch processing)

Um processamento em lote (batch processing) é usado quando uma _goroutine_ gera itens um por um, mas o consumidor deseja processar os itens em blocos. Normalmente, um canal de conclusão é usado para notificar o escritor que o item foi processado. Um canal de descarga pode ser usado para forçar que o buffer seja enviado antes que ele esteja cheio.

Exemplo: Ao invés de salvar cada item no banco de dados assim que ele é recebido, é possível utilizar um buffer de 100 itens ou 100ms e salvar os itens em uma única requisição.

No exemplo, quando a terceira requisição (req) é enviada, o buffer percebe que ele está cheio e envia os dados para o canal de saída.

Há um canal que permite enviar os dados antes que o buffer esteja cheio, chamado `descarga`.

Quando o canal de entrada é fechado, mas ainda há itens no buffer, o buffer é enviado para o canal de saída.

```go
package main

import (
	"fmt"
)

type req struct {
	valor int
}

func processar(lote []req) {
	fmt.Println("processando lote com valores: ", lote)
}

func processadorLotes(entrada <-chan []req) chan bool {
	pronto := make(chan bool)
	go func() {
		for lote := range entrada {
			processar(lote)
		}
		pronto <- true
	}()
	return pronto
}

func processamentoLotes(entrada <-chan req, descarga <-chan struct{}, tamanhoLote int) chan []req {
	saida := make(chan []req)
	go func() {
		defer close(saida)
		buf := make([]req, 0, tamanhoLote)

		for {
			select {
			// enquanto houver itens para processar
			case r, ok := <-entrada:
				if !ok {
					// envia o que tiver no buffer antes de sair
					if len(buf) > 0 {
						saida <- buf
					}
					// para o loop quando o canal de entrada for fechado
					return
				}
				// Adiciona o item no buffer
				buf = append(buf, r)
				// se o buffer estiver cheio, descarrega
				if len(buf) == tamanhoLote {
					saida <- buf
					buf = make([]req, 0, tamanhoLote)
				}

			// Se receber um sinal de descarga, descarrega o que tiver no buffer
			case <-descarga:
				if len(buf) > 0 {
					saida <- buf
					buf = make([]req, 0, tamanhoLote)
				}
			}
		}
	}()
	return saida
}

func main() {
	entrada := make(chan req)
	descarga := make(chan struct{})

	// inicia de forma concorrente o processamento em lotes
	saida := processamentoLotes(entrada, descarga, 3)
	// O consumidor de lotes será iniciado de forma concorrente
	pronto := processadorLotes(saida)

	entrada <- req{valor: 1}
	entrada <- req{valor: 2}
	entrada <- req{valor: 3}

	// Envia mais dois itens, porém força a descarga
	// através de um sinal
	entrada <- req{valor: 4}
	entrada <- req{valor: 5}
	descarga <- struct{}{}

	// Envia mais dois itens, não o suficiente para descarregar
	// o lote.
	entrada <- req{valor: 6}
	entrada <- req{valor: 7}
	// Eles serão processados mesmo assim.

	close(entrada)

	// Aguarda todo o processamento do processador de lotes
	// antes de encerrar o programa
	<-pronto
}
```

## 🎫 Sistema de ticket

Um sistema de ticket é usado para controlar quando um determinado trabalho pode ser executado, normalmente é utilizado para limitar o uso de um recurso sobre um período de tempo.

Exemplo: Uma API pode ser acionada apenas 15 vezes em um período de 15 minutos. A utilização é medida em blocos de 15 minutos.

No exemplo, a bilheteria é um sistema de ticket que garante que apenas 15 "tickets" sejam processados a cada segundo.

Enviamos através de um canal 30 processamentos a serem feitos, mas o sistema de ticket garante que apenas 15 processamentos sejam executados por segundo.

Como pode ser visto, o trabalhador fica bloqueado até que um ticket seja enviado através do canal.

```go
package main

import (
	"fmt"
	"time"
)

type (
	Trabalho func()
	ticket   int
)

func trabalhador(tickets <-chan ticket, work <-chan Trabalho) {
	for w := range work {
		<-tickets // espera por um ticket
		w()       // executa um trabalho
	}
}

func bilheteria(tickets chan<- ticket, timeout time.Duration, nTickets int) {
	for {
		for i := range nTickets {
			tickets <- ticket(i)
		}

		// espera até que mais tickets possam ser emitidos
		<-time.After(timeout)
	}
}

func main() {
	tickets := make(chan ticket)
	trabalhos := make(chan Trabalho)
	pronto := make(chan struct{})

	go bilheteria(tickets, 1*time.Second, 10)
	go func() {
		trabalhador(tickets, trabalhos)
		pronto <- struct{}{}
	}()

	for i := 0; i <= 30; i++ {
		trabalhos <- func() {
			fmt.Println("processando ticket")
		}
		fmt.Println("trabalho ", i, " enviado")
	}

	close(trabalhos)
	<-pronto
}
```
