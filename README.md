# Concorrência em Go

Go é fundamentada no modelo CSP (Communicating sequential processes) proposto por Tony Hoare. Neste modelo, os dados são compartilhados enviando mensagens por canais.

No CSP original de Hoare, um processo envia mensagens direto para outro, identificado pelo nome. Erlang seguiu esse caminho. Go vem de outro ramo da família, o das linguagens Newsqueak, Alef e Limbo. Nelas o canal é um valor de primeira classe, que pode ser guardado em variáveis, passado como parâmetro e até enviado por outro canal. Os dois modelos são equivalentes, só se expressam de forma diferente. Rob Pike compara com arquivos: em Erlang é como escrever em um arquivo pelo nome, em Go é como escrever por meio de um descritor de arquivo.

Outra ideia que volta várias vezes é que [concorrência não é paralelismo](https://go.dev/blog/waza-talk). Concorrência é compor computações independentes, uma forma de estruturar o programa. Paralelismo é executá-las ao mesmo tempo. Um programa concorrente pode rodar em um único processador. E um programa bem estruturado para concorrência costuma paralelizar bem quando há mais de um.

As explicações e exemplos vêm em boa parte da [apresentação](https://github.com/andrebq/andrebq.github.io) do @andrebq.

Outras influências:

- O [artigo](https://go.dev/blog/pipelines) sobre _pipelines_ e cancelamento em Go.
- A palestra [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide) de Rob Pike (Google I/O 2012), de onde vêm os geradores, o fan-in, os timeouts com `select` e o canal de parada.
- A palestra [Advanced Go Concurrency Patterns](https://go.dev/talks/2013/advconc.slide) de Sameer Ajmani (Google I/O 2013), de onde vêm o laço `for` com `select` e estado local, a parada confirmada por canal de resposta e o canal `nil` no `select`. As duas palestras são anteriores ao pacote `context`, e a abertura da [Parte 3](#parte-3--encerrando-goroutines) diz o que mudou com ele.
- Os [Go Proverbs](https://go-proverbs.github.io/), também de Rob Pike (Gopherfest 2015): "_Don't communicate by sharing memory, share memory by communicating_", "_Concurrency is not parallelism_", "_Channels orchestrate; mutexes serialize_" e "_Clear is better than clever_".

Cada parte vai do mais simples ao mais complexo. A [Parte 1](#parte-1--fundamentos) apresenta as peças da linguagem: canais, `select` e `WaitGroup`. A [Parte 2](#parte-2--padrões-básicos) traz os padrões básicos. A [Parte 3](#parte-3--encerrando-goroutines) ensina a encerrar _goroutines_. A [Parte 4](#parte-4--controlando-o-ritmo) trata de produtores e consumidores em ritmos diferentes. A [Parte 5](#parte-5--padrões-avançados) reúne as combinações. Para ir mais fundo em `context`, sugiro também [este repositório](https://github.com/cassiobotaro/contexto).

Os `fmt.Print` dos exemplos estão ali só para você enxergar a execução e não fazem parte dos padrões. Em código real seriam _logs_, ou nem existiriam.

Cada pasta é um programa independente, que você executa com `go run ./pipeline/`. Cada uma tem um `exemplo_test.go` com uma função `Example` que confere a saída. Ela usa `// Output:` quando a ordem das linhas é fixa e `// Unordered output:` quando só o conjunto é previsível. Para rodar tudo com o detector de corrida, use `go test -race ./...`.

Os mesmos padrões aparecem com outros nomes em livros, artigos e outras linguagens, por isso cada seção traz uma linha "Também conhecido como". Dois pedem cuidado: "produtor" e "consumidor" são papéis, não padrões, e quase todo exemplo tem os dois. Aparecem como nomes alternativos de [Geradores](#-geradores) e [Trabalhador](#-trabalhador-worker) porque nessas seções cada papel aparece sozinho.

## 📑 Sumário

- [Parte 1 · Fundamentos](#parte-1--fundamentos)
  - [🔗 Canais](#-canais)
  - [🗺️ Olá Mundo](#️-olá-mundo)
  - [🎛️ Select e timeouts](#️-select-e-timeouts)
  - [⏳ Esperando goroutines (WaitGroup)](#-esperando-goroutines-waitgroup)
- [Parte 2 · Padrões básicos](#parte-2--padrões-básicos)
  - [🆕 Geradores](#-geradores)
  - [🚧 Trabalhador (worker)](#-trabalhador-worker)
  - [🏭 Pipeline](#-pipeline)
  - [📣 Fan-out](#-fan-out)
  - [🔀 Tee (broadcast)](#-tee-broadcast)
    - [Tee com timeout](#tee-com-timeout)
  - [⚗️ Fan-in](#️-fan-in)
    - [Fan-in com uma _goroutine_ e `select`](#fan-in-com-uma-goroutine-e-select)
  - [👷 Grupo de Trabalhadores (pool of workers)](#-grupo-de-trabalhadores-pool-of-workers)
  - [📨 Requisição e resposta](#-requisição-e-resposta)
- [Parte 3 · Encerrando goroutines](#parte-3--encerrando-goroutines)
  - [🚏 Canal de parada (quit channel)](#-canal-de-parada-quit-channel)
  - [🛑 Vazamento de goroutines e context](#-vazamento-de-goroutines-e-context)
  - [🤝 Parada com confirmação](#-parada-com-confirmação)
  - [🧩 Combinar sinais de parada (or-channel)](#-combinar-sinais-de-parada-or-channel)
- [Parte 4 · Controlando o ritmo](#parte-4--controlando-o-ritmo)
  - [🚦 Contrapressão (backpressure)](#-contrapressão-backpressure)
  - [🚥 Semáforo (paralelismo limitado)](#-semáforo-paralelismo-limitado)
    - [E com errgroup?](#e-com-errgroup)
  - [🎫 Sistema de ticket](#-sistema-de-ticket)
  - [📦 Processamento em lote (batch processing)](#-processamento-em-lote-batch-processing)
  - [🪟 Janela deslizante](#-janela-deslizante)
- [Parte 5 · Padrões avançados](#parte-5--padrões-avançados)
  - [🔐 Goroutine dona do estado](#-goroutine-dona-do-estado)
    - [E com mutex?](#e-com-mutex)
  - [🏁 Primeiro a responder](#-primeiro-a-responder)
  - [💓 Heartbeat](#-heartbeat)
- [Curiosidade](#curiosidade)
  - [⛓️ Daisy-chain](#️-daisy-chain)

## Parte 1 · Fundamentos

Antes dos padrões, as peças da linguagem que todos eles usam. Esta parte não apresenta nenhum padrão. Ela explica os canais, o `select` e a forma de esperar _goroutines_ terminarem. Se você já conhece essas peças, pode ir direto para a Parte 2.

### 🔗 Canais

Canais (_channels_) são uma primitiva da linguagem. Você os usa para enviar e receber valores entre _goroutines_. Os valores podem ser de qualquer tipo, inclusive do tipo canal.

Um canal é um ponto de sincronização entre _goroutines_. Quem escreve fica bloqueado até alguém ler, e quem lê fica bloqueado até alguém escrever ou fechar o canal.

Fechar um canal indica que nenhum outro valor será escrito nele. Daí em diante, ler devolve o valor zero do tipo, e escrever causa um erro em tempo de execução (_panic_).

**Fechar como sinal.** Fechar um canal é a forma usual em Go de comunicar um evento que acontece uma única vez, como "terminei" ou "pode parar". Funciona para qualquer número de leitores, porque todos os que estiverem lendo são desbloqueados ao mesmo tempo. Para isso costuma-se usar um `chan struct{}`, que não carrega dado nenhum. O primeiro uso aparece no [trabalhador](#-trabalhador-worker).

**Canais com buffer.** `make(chan int, 3)` cria um canal que guarda até três valores. Enviar só bloqueia quando o buffer está cheio, e receber só bloqueia quando ele está vazio. O buffer tira a sincronização entre quem envia e quem recebe, e por isso pede mais cuidado. Os exemplos daqui usam canais sem buffer sempre que podem. O buffer só aparece quando ele é a própria ideia do padrão, como na [contrapressão](#-contrapressão-backpressure), no [semáforo](#-semáforo-paralelismo-limitado) e no [primeiro a responder](#-primeiro-a-responder).

**Direção.** Na assinatura de uma função, um canal pode ser declarado só para leitura (`<-chan int`) ou só para escrita (`chan<- int`). Com isso o compilador impede que a função leia do canal em que só deveria escrever, ou escreva naquele em que só deveria ler. Todos os exemplos usam tipos direcionais nas assinaturas. A única exceção é o canal `quit` da [parada com confirmação](#-parada-com-confirmação), usado nos dois sentidos de propósito.

### 🗺️ Olá Mundo

O [primeiro exemplo](./ola_mundo/ola_mundo.go) mostra como criar um canal, que serve de ponte entre a função principal e uma _goroutine_.

O programa principal fica bloqueado em `<-canal` até que a _goroutine_ envie a mensagem "Olá, mundo!".

Quando isto ocorre, o programa principal é desbloqueado, recebe a mensagem e a exibe. Como o canal não tem buffer, o envio e o recebimento acontecem juntos.

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

### 🎛️ Select e timeouts

O `select` é uma estrutura de controle feita para concorrência. Ele parece um `switch`, mas cada `case` é uma comunicação: um envio ou um recebimento em um canal. O `select` bloqueia até que alguma das comunicações possa prosseguir. Se várias puderem ao mesmo tempo, ele escolhe uma de forma pseudoaleatória. Se houver um `default`, ele não bloqueia. Na palestra [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide), Rob Pike diz que o `select` é a razão de canais e _goroutines_ fazerem parte da linguagem, em vez de serem uma biblioteca.

Quase todos os padrões das próximas partes dependem dele. O uso mais comum é dar um prazo a uma comunicação. Os dois exemplos usam o mesmo gerador, `tagarela`, que fala cada vez mais devagar:

- **Timeout por mensagem.** `time.After` devolve um canal que recebe um valor depois do tempo indicado. Colocado em um `case` dentro do laço, ele é recriado a cada volta. O prazo vale para cada mensagem e é renovado sempre que uma chega.
- **Timeout para a conversa inteira.** É o mesmo `time.After`, mas criado uma única vez, fora do laço. O prazo vale para a conversa toda, não importa quantas mensagens cheguem.

Até o Go 1.22, cada `time.After` deixava um timer vivo até disparar, mesmo depois de o `select` ter escolhido outro `case`. Com um prazo de 350ms e mensagens a cada 100ms, sobravam timers esperando à toa, e a recomendação era evitar `time.After` dentro de laço. Desde o Go 1.23, um timer que o programa não referencia mais é recolhido pelo coletor de lixo na hora, desde que o `go.mod` declare `go 1.23` ou mais novo, como o deste repositório. O timeout por mensagem, do jeito que está, é a forma correta. Em código real o prazo costuma chegar de fora, em um `context.Context` criado com `context.WithTimeout`, e o `case` passa a ser `<-ctx.Done()`. Esse é o timeout da conversa inteira, com a diferença de que o mesmo prazo atravessa as funções chamadas. A [Parte 3](#parte-3--encerrando-goroutines) mostra isso.

Dois outros recursos do `select` aparecem mais adiante. O `default`, que torna o `select` não bloqueante, é usado na [contrapressão](#-contrapressão-backpressure) e no [heartbeat](#-heartbeat). O outro é o canal `nil`. Um `select` nunca escolhe um `case` cujo canal é `nil`, então atribuir `nil` à variável do canal desliga aquele `case` enquanto o laço continua rodando. Esse truque aparece no [fan-in com select](#fan-in-com-uma-goroutine-e-select) e na [janela deslizante](#-janela-deslizante). Ele é uma das três técnicas da palestra [Advanced Go Concurrency Patterns](https://go.dev/talks/2013/advconc.slide), de Sameer Ajmani (2013), que tem um slide só para ele.

O `tagarela` recebe um canal `quit`, que a função fecha ao sair para que o gerador também termine. Esse é o [canal de parada](#-canal-de-parada-quit-channel), assunto da Parte 3.

```go
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
```

### ⏳ Esperando goroutines (WaitGroup)

No [Olá Mundo](#️-olá-mundo), o programa principal esperou a _goroutine_ lendo de um canal. Quando o que se quer é só esperar várias _goroutines_ terminarem, sem receber nenhum dado delas, a ferramenta certa é o `sync.WaitGroup`.

São duas chamadas. O `wg.Go` dispara a função em uma nova _goroutine_ e registra no `WaitGroup` que ela precisa terminar. O `wg.Wait()` bloqueia até que todas as _goroutines_ disparadas assim terminem.

No exemplo, três tarefas com durações diferentes são disparadas de uma vez. A função principal só imprime a última linha depois que as três terminaram. Experimente comentar o `wg.Wait()` e veja o programa acabar antes de qualquer tarefa imprimir.

O `wg.Go` existe desde o Go 1.25. Em código mais antigo você vai encontrar a forma equivalente, com `wg.Add(1)` antes de cada `go` e `defer wg.Done()` dentro da _goroutine_. O `wg.Go` faz as duas coisas de uma vez e evita os erros clássicos dessa forma, como esquecer o `Done` ou chamar o `Add` dentro da _goroutine_.

Por que um `WaitGroup` e não um canal? Canais servem para orquestrar o fluxo de dados entre _goroutines_. Contar quantas já terminaram é um problema menor, e para esses Rob Pike recomenda o pacote `sync`. Na palestra [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide) ele avisa "_Don't overdo it_", porque às vezes só é preciso um contador. É o provérbio "_Channels orchestrate; mutexes serialize_", dos [Go Proverbs](https://go-proverbs.github.io/). Por isso, nos padrões a seguir, os canais transportam os dados e o `WaitGroup` apenas conta quem terminou.

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

// tarefa simula um trabalho que leva algum tempo.
func tarefa(id int) {
	time.Sleep(time.Duration(id) * 50 * time.Millisecond)
	fmt.Printf("tarefa %d terminou\n", id)
}

func main() {
	var wg sync.WaitGroup

	// wg.Go dispara a função em uma nova goroutine e registra no
	// WaitGroup que ela precisa terminar.
	for i := range 3 {
		wg.Go(func() {
			tarefa(i + 1)
		})
	}

	// Bloqueia até que todas as goroutines disparadas com wg.Go terminem.
	// Sem esta linha o programa acabaria antes de as tarefas imprimirem.
	wg.Wait()
	fmt.Println("todas as tarefas terminaram")
}
```

> **É possível fazer só com canais.** Um canal com buffer de tamanho `n` e um laço que lê `n` vezes fazem o mesmo papel. Cada tarefa envia um sinal ao terminar, e a função principal espera receber todos. Funciona, mas é um `WaitGroup` refeito à mão.
>
> ```go
> terminar := make(chan struct{}, 3)
>
> for i := range 3 {
> 	go func() {
> 		tarefa(i + 1)
> 		terminar <- struct{}{}
> 	}()
> }
>
> for range 3 {
> 	<-terminar
> }
> ```

## Parte 2 · Padrões básicos

Os blocos de montar. Cada padrão desta parte faz uma coisa só e usa apenas o que foi visto nos fundamentos. Os padrões das partes seguintes são combinações e variações destes.

### 🆕 Geradores

**Também conhecido como:** produtor, _source_. É o mesmo papel do `produtor` da seção de [contrapressão](#-contrapressão-backpressure).

Um gerador é uma função que dispara uma _goroutine_ para escrever uma sequência de valores em um canal, e devolve esse canal a quem a chamou. O produtor roda em paralelo com o consumidor. Isso importa quando produzir custa caro, como ler de disco ou de rede, e quando os valores vão atravessar um [_pipeline_](#-pipeline) de etapas concorrentes.

No [exemplo](./geradores/geradores.go), `sequenciaNumeros` gera mil inteiros. A função principal lê o canal e imprime os valores. Com o `range`, a iteração continua até o canal ser fechado.

Repare no `context.Context` e no `select` dentro da _goroutine_. Cada envio disputa com `ctx.Done()`. Sem isso, um consumidor que parasse de ler no meio deixaria a _goroutine_ presa para sempre no próximo envio, com o canal e tudo o que ela segura. Com o contexto, quem consome cancela, a _goroutine_ sai e fecha o canal ao sair. No `main` isso não chega a acontecer, porque o laço lê os mil valores. O `exemplo_test.go` da pasta faz o outro caminho: lê um valor, cancela e drena o canal até ele fechar, o que prova que a _goroutine_ terminou. É a forma de escrever um gerador em código de hoje, e o mecanismo por trás dela é o assunto da [Parte 3](#parte-3--encerrando-goroutines).

O custo é que a responsabilidade passa para quem consome. Ele precisa criar o contexto e cancelar ao sair, mesmo quando leu tudo, e o `defer cancelar()` do exemplo está ali por isso. Esquecer o `cancel` é um erro que o `go vet` aponta (`lostcancel`): o contexto filho fica registrado no pai até o pai ser cancelado.

A função `sequenciaNumeros` reaparece em vários exemplos, sem o contexto, para que cada arquivo fique no assunto da própria seção. Ela é copiada de propósito, para que cada um possa ser lido e executado sozinho. Como diz um dos [Go Proverbs](https://go-proverbs.github.io/), "_a little copying is better than a little dependency_".

> **Atenção:** as cópias sem contexto dos outros exemplos vazam a _goroutine_ se o consumidor parar de ler antes do fim. Cada seção avisa quando isso acontece.

```go
package main

import (
	"context"
	"fmt"
)

// sequenciaNumeros gera os inteiros de inicial a final em uma goroutine e
// os envia por um canal. Cada envio disputa com ctx.Done(): se o consumidor
// cancelar o contexto, a goroutine sai em vez de ficar presa no envio.
func sequenciaNumeros(ctx context.Context, inicial, final int) <-chan int {
	saida := make(chan int)
	go func() {
		// fecha o canal ao sair, tanto no fim quanto no cancelamento
		defer close(saida)
		for i := inicial; i <= final; i++ {
			select {
			case saida <- i:
			case <-ctx.Done():
				return
			}
		}
	}()
	return saida
}

func main() {
	ctx, cancelar := context.WithCancel(context.Background())
	defer cancelar()

	for valor := range sequenciaNumeros(ctx, 1, 1000) {
		fmt.Printf("valor: %v\n", valor)
	}
}
```

### 🚧 Trabalhador (worker)

**Também conhecido como:** consumidor, _sink_. O segundo só vale quando o trabalhador é o último estágio, isto é, quando não repassa nada adiante.

Um trabalhador é uma _goroutine_ que recebe valores de um canal e os processa.

No exemplo, a função principal envia dez valores inteiros pelo canal de entrada, e um trabalhador os processa.

Vários trabalhadores podem ler do mesmo canal. É o [fan-out](#-fan-out), mais adiante.

Repare que o término é sinalizado com `close(pronto)`, e não com o envio de um valor. É o fechamento usado como sinal, visto em [canais](#-canais).

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
		// Fechar o canal é o idioma para sinalizar um evento único:
		// comunica "terminou" a qualquer número de leitores.
		close(pronto)
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

### 🏭 Pipeline

**Também conhecido como:** cadeia de estágios. Cada função do _pipeline_ é um _estágio_ (_stage_).

Um _pipeline_ recebe valores de um canal e escreve em outro, normalmente depois de transformar o valor.

No exemplo, a função `dobro` é um estágio: lê os valores do canal de entrada e escreve os valores dobrados no canal de saída.

Repare nas assinaturas. A função `dobro` recebe um `<-chan int` e devolve outro `<-chan int`. São os tipos direcionais vistos em [canais](#-canais), e aqui eles deixam claro quem lê e quem escreve em cada estágio.

Os valores gerados por `sequenciaNumeros` são enviados para o canal de entrada do _pipeline_. A função principal recebe os valores transformados pelo canal de saída e os imprime.

Os estágios podem ser encadeados. No exemplo, `dobro` é aplicado duas vezes, e cada valor sai multiplicado por quatro.

> **Atenção:** o gerador e as etapas deste _pipeline_ não são canceláveis, e vazam se o consumidor parar de ler antes do fim. Veja a [Parte 3](#parte-3--encerrando-goroutines).

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

### 📣 Fan-out

**Também conhecido como:** distribuição, _work distribution_.

Um fan-out distribui os valores de um canal de entrada entre várias _goroutines_. O artigo sobre [_pipelines_](https://go.dev/blog/pipelines) define assim: múltiplas funções lendo do mesmo canal até que ele seja fechado. Cada valor é processado por exatamente uma delas, o que permite dividir um trabalho demorado entre vários trabalhadores.

Não é preciso nenhum código para decidir quem recebe o quê, porque o próprio canal faz a distribuição. Quando várias _goroutines_ estão bloqueadas lendo o mesmo canal, cada envio é entregue a apenas uma delas.

No exemplo, três trabalhadores dividem entre si os dez valores gerados por `sequenciaNumeros`. Repare na saída que nenhum valor aparece duas vezes. Um [`sync.WaitGroup`](#-esperando-goroutines-waitgroup) aguarda o término de todos. O [grupo de trabalhadores](#-grupo-de-trabalhadores-pool-of-workers), mais adiante, é uma aplicação deste padrão.

Execute o exemplo mais de uma vez e veja que a ordem da saída muda. Os trabalhadores concorrem pelos valores da entrada, e quem decide qual deles roda a cada momento é o escalonador. É a primeira vez que o não determinismo aparece por aqui, e ele vem da ideia de que [concorrência não é paralelismo](https://go.dev/blog/waza-talk): o programa descreve computações independentes, mas não diz em que ordem elas executam. Por isso um programa concorrente correto não pode depender dessa ordem.

Não confunda com o [tee](#-tee-broadcast), em que cada valor é copiado para todos os consumidores.

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

// trabalhador lê do canal de entrada, que é compartilhado com os demais
// trabalhadores. Cada valor é entregue a exatamente um deles: quem estiver
// livre primeiro, recebe.
func trabalhador(id int, entrada <-chan int) {
	for valor := range entrada {
		fmt.Printf("id: %d processando valor: %v\n", id, valor)
		// Simula um processamento demorado
		time.Sleep(100 * time.Millisecond)
	}
}

// fanout distribui os valores de um único canal de entrada entre n
// trabalhadores e só retorna quando todos terminarem.
func fanout(entrada <-chan int, n int) {
	var wg sync.WaitGroup

	for i := range n {
		wg.Go(func() {
			trabalhador(i+1, entrada)
		})
	}
	wg.Wait()
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
	// Três trabalhadores dividem entre si os dez valores da sequência
	fanout(sequenciaNumeros(1, 10), 3)
}
```

### 🔀 Tee (broadcast)

**Também conhecido como:** _broadcast_, _publish/subscribe_ em memória. O segundo é aproximado: em um _pub/sub_ os assinantes costumam entrar e sair dinamicamente, enquanto o tee tem um conjunto fixo de saídas.

Um tee copia cada valor de um canal de entrada para todos os canais de saída, de modo que todos os consumidores veem todos os valores. O nome vem do comando `tee` do Unix, que duplica o que recebe. É o oposto do [fan-out](#-fan-out), em que cada valor vai para um único consumidor.

No exemplo, uma sequência de dez números é copiada para dois canais de saída. Cada canal tem seu trabalhador, e os dois recebem todos os valores.

O tee lê cada valor da entrada e o envia, em sequência, para cada uma das saídas. Quando a entrada é fechada, ele fecha todas as saídas. Para aguardar o término dos trabalhadores, a função principal usa um [`sync.WaitGroup`](#-esperando-goroutines-waitgroup).

Como os canais não têm buffer, o tee só passa para o próximo valor depois que todas as saídas receberam o atual. A consequência é que um consumidor lento atrasa todos os outros, e também o produtor. É a [contrapressão](#-contrapressão-backpressure) aplicada ao broadcast. Ninguém perde mensagem, mas todos andam no ritmo do mais lento.

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

// tee copia cada valor da entrada para todas as saídas: todos os consumidores
// veem todos os valores. O envio é sequencial e sem buffer, então o tee só
// avança quando todas as saídas receberam o valor: um consumidor lento
// atrasa todos os outros.
func tee(entrada <-chan int, saidas ...chan<- int) {
	for valor := range entrada {
		for _, saida := range saidas {
			saida <- valor
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

// trabalhador consome os valores de uma das saídas do tee. O parâmetro
// `demora` simula o tempo de processamento de cada valor.
func trabalhador(id int, entrada <-chan int, demora time.Duration) {
	for valor := range entrada {
		fmt.Println("id: ", id, " valor: ", valor)
		time.Sleep(demora)
	}
}

func main() {
	saida1 := make(chan int)
	saida2 := make(chan int)

	// Aguarda o término dos trabalhadores
	var wg sync.WaitGroup
	wg.Go(func() { trabalhador(1, saida1, 0) })
	wg.Go(func() { trabalhador(2, saida2, 0) })

	// Copia a sequência de números para todos os canais de saída
	tee(sequenciaNumeros(1, 10), saida1, saida2)
	wg.Wait()

	// Tee com timeout (veja tee_timeout.go): agora o trabalhador 2 é mais lento
	// do que o timeout, então parte dos valores destinados a ele é descartada.
	saida1 = make(chan int)
	saida2 = make(chan int)

	wg.Go(func() { trabalhador(1, saida1, 0) })
	wg.Go(func() { trabalhador(2, saida2, 250*time.Millisecond) })

	teeComTimeout(sequenciaNumeros(1, 5), 100*time.Millisecond, saida1, saida2)
	wg.Wait()
}
```

#### Tee com timeout

Se um consumidor lento não pode segurar os demais, uma alternativa é desistir do envio depois de um tempo. [Nesta variante](./tee/tee_timeout.go), cada envio é feito dentro de um `select` que disputa com `time.After`, e vence o que acontecer primeiro. Se o tempo esgotar, o valor é descartado apenas para aquela saída e o tee segue em frente. Um `select` por saída dentro do laço é suficiente, não é preciso criar uma _goroutine_ para cada envio.

Descartar mensagens é uma decisão de projeto, e não parte do padrão. Com o descarte, o consumidor lento deixa de ver todos os valores, que era justamente a garantia do tee. Por isso o exemplo avisa na saída cada vez que descarta, em vez de descartar em silêncio. O timeout limita o atraso, mas não acaba com ele. Cada valor ainda pode esperar até `timeout` em cada saída lenta. Outras formas de lidar com um consumidor lento aparecem na [janela deslizante](#-janela-deslizante) e na [contrapressão](#-contrapressão-backpressure).

No exemplo, a função principal executa as duas versões. Na segunda, o trabalhador 2 leva 250ms por valor e o timeout é de 100ms, então parte dos valores destinados a ele é descartada.

```go
package main

import (
	"fmt"
	"time"
)

// teeComTimeout é um tee que não espera indefinidamente por um consumidor
// lento: se uma saída não receber o valor dentro de `timeout`, o valor é
// descartado para aquela saída e o tee segue em frente.
// Descartar mensagens é uma decisão de projeto, não parte do padrão.
func teeComTimeout(entrada <-chan int, timeout time.Duration, saidas ...chan<- int) {
	for valor := range entrada {
		for i, saida := range saidas {
			// Um select por saída: o que acontecer primeiro, o envio ou o timeout
			select {
			case saida <- valor:
			case <-time.After(timeout):
				// Sinalizamos o descarte explicitamente para não perder
				// a informação silenciosamente.
				fmt.Printf("tee: descarte por timeout, saida=%d valor=%d\n", i+1, valor)
			}
		}
	}
	// Como a entrada foi consumida, fecha os canais de saída
	for _, saida := range saidas {
		close(saida)
	}
}
```

### ⚗️ Fan-in

**Também conhecido como:** _merge_, multiplexação (o termo que Rob Pike usa na palestra de 2012).

Um fan-in copia dados de múltiplos canais de entrada e escreve em um único canal de saída. Normalmente um fan-in só termina quando todos os canais de entrada são fechados.

A função fan-in recebe os canais de entrada como [parâmetros múltiplos](https://gobyexample.com/variadic-functions).

No exemplo, três geradores são passados para a função fan-in, que devolve um único canal de saída. Por dentro há uma _goroutine_ por canal de entrada, e todas escrevem no mesmo canal de saída.

Escrever em um canal fechado causa um _panic_, então a saída só pode ser fechada depois que todas as entradas terminarem. Um [`sync.WaitGroup`](#-esperando-goroutines-waitgroup) conta quantas ainda faltam.

Repare na _goroutine_ que espera em `wg.Wait()` e fecha a saída quando a última entrada acaba.

> **Atenção:** estes geradores não são canceláveis, e vazam se o consumidor parar de ler antes do fim. Veja a [Parte 3](#parte-3--encerrando-goroutines).

```go
package main

import (
	"fmt"
	"sync"
)

// fanin combina vários canais de entrada em um único canal de saída.
// Utiliza um WaitGroup para saber quando todos os canais de entrada foram processados.
func fanin(entradas ...<-chan int) <-chan int {
	saida := make(chan int)
	var wg sync.WaitGroup

	for _, entrada := range entradas {
		// Uma goroutine por entrada; o WaitGroup é avisado quando ela termina
		wg.Go(func() {
			for valor := range entrada {
				saida <- valor
			}
		})
	}

	// Quando todos os canais de entrada terminarem, fecha o canal de saída
	go func() {
		wg.Wait()
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

	// Com um número fixo de entradas, uma única goroutine com select basta
	// (veja fan_in_select.go)
	canal = faninSelect(
		sequenciaNumeros(31, 40),
		sequenciaNumeros(41, 50),
	)
	for valor := range canal {
		fmt.Printf("valor (select): %v\n", valor)
	}
}
```

#### Fan-in com uma _goroutine_ e `select`

Quando o número de entradas é fixo e conhecido, Rob Pike mostra na palestra [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide) uma variante mais enxuta. Uma única _goroutine_ com um `select` repassa para a saída o valor da entrada que estiver pronta primeiro.

A versão da palestra roda para sempre. [Aqui](./fan_in/fan_in_select.go) ela também trata o fechamento das entradas, com o truque do canal `nil` visto em [select](#️-select-e-timeouts). Quando uma entrada é fechada, a variável vira `nil` e aquele `case` deixa de ser escolhido. Quando todas viram `nil`, o laço termina e a saída é fechada. Como só uma _goroutine_ escreve na saída, ela mesma fecha o canal, sem `WaitGroup`.

Quando usar cada uma? Se o número de canais é variável, como em um _slice_ ou em parâmetros múltiplos, use uma _goroutine_ por entrada, porque um `select` tem um número fixo de `case`s escrito no código. Se o número é fixo e pequeno, o `select` é mais direto. Basta uma _goroutine_, e não é preciso contar quem terminou.

```go
package main

// faninSelect combina um número fixo de canais de entrada (aqui, dois) usando
// uma única goroutine e um select, em vez de uma goroutine por entrada.
// Como só uma goroutine escreve na saída, ela mesma fecha o canal ao terminar:
// não é preciso contar ninguém.
func faninSelect(entrada1, entrada2 <-chan int) <-chan int {
	saida := make(chan int)
	go func() {
		defer close(saida)
		for entrada1 != nil || entrada2 != nil {
			select {
			case valor, ok := <-entrada1:
				if !ok {
					// Entrada fechada: um canal nil nunca é selecionado,
					// o que desabilita este case.
					entrada1 = nil
					continue
				}
				saida <- valor
			case valor, ok := <-entrada2:
				if !ok {
					entrada2 = nil
					continue
				}
				saida <- valor
			}
		}
	}()
	return saida
}
```

### 👷 Grupo de Trabalhadores (pool of workers)

**Também conhecido como:** _worker pool_, _pool_ de _goroutines_.

A piscina de marmotinhas (carinhosamente chamada pela minha esposa) é uma coleção de _goroutines_ que ficam esperando tarefas serem atribuídas a elas. Quando termina uma tarefa, a _goroutine_ volta a ficar disponível para a próxima.

No exemplo, dois trabalhadores esperam valores no canal de entrada. Cada um dobra o valor que recebe e envia o resultado pelo canal de saída.

O grupo de trabalhadores é uma aplicação de [fan-out](#-fan-out). Várias _goroutines_ leem do mesmo canal e cada valor vai para uma só. Além de distribuir o trabalho, o grupo junta os resultados em um canal de saída.

O grupo fixa quantas _goroutines_ existem. Se a ideia for ter uma _goroutine_ por tarefa e limitar apenas quantas executam ao mesmo tempo, veja o [semáforo](#-semáforo-paralelismo-limitado). Se as tarefas devolvem erro, nos dois casos, veja a variante [com errgroup](#e-com-errgroup).

Como no [fan-out](#-fan-out), a ordem da saída muda a cada execução.

Os trabalhadores são iniciados com `wg.Go`, e outra _goroutine_ espera em `wg.Wait()` para fechar o canal de saída, como em [esperando goroutines](#-esperando-goroutines-waitgroup). Repare que o `trabalhador` nem sabe que o `WaitGroup` existe. Ele só processa valores, e quem o dispara é que cuida de esperar.

```go
package main

import (
	"fmt"
	"sync"
)

// trabalhador processa valores recebidos do canal de entrada e envia resultados para o canal de saída.
func trabalhador(id int, entrada <-chan int, saida chan<- int) {
	for valor := range entrada {
		fmt.Printf("id: %d processou valor: %v\n", id, valor)
		saida <- valor * 2
	}

	fmt.Printf("id: %d terminou\n", id)
}

func grupoDeTrabalhadores(entrada <-chan int, nTrabalhadores int) <-chan int {
	saida := make(chan int)
	// Os canais transportam os dados; o WaitGroup apenas conta
	// quantos trabalhadores ainda não terminaram.
	var wg sync.WaitGroup

	// Cria e inicia os trabalhadores. wg.Go dispara a função em uma nova
	// goroutine e registra no WaitGroup que ela precisa terminar.
	for i := range nTrabalhadores {
		wg.Go(func() {
			trabalhador(i+1, entrada, saida)
		})
	}

	// Goroutine para fechar o canal de saída quando todos os trabalhadores terminarem
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

### 📨 Requisição e resposta

**Também conhecido como:** canal de resposta, _RPC_ interno, _restoring sequencing_ (o nome que Rob Pike dá a um uso específico da ideia, comentado abaixo).

Canais são valores como qualquer outro, então uma mensagem pode carregar um canal. Quem envia uma requisição inclui nela o canal pelo qual quer receber a resposta e fica bloqueado lendo desse canal. Quem atende processa e responde no canal que veio na mensagem. Nenhum estado é compartilhado, pois pedido e resposta viajam por canais. É assim que se faz uma _goroutine_ funcionar como um serviço, e a ideia reaparece na [goroutine dona do estado](#-goroutine-dona-do-estado).

No exemplo, a função principal envia cinco requisições ao `servico` e espera cada resposta antes de enviar a próxima. O campo `resposta` é declarado como `chan<- int`, então o serviço só pode escrever nele.

Na palestra [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide), Pike usa a mesma ideia para "restaurar a sequência" de um fan-in. Cada mensagem carrega um canal `wait`, e quem produziu só envia a próxima depois que o leitor avisa, por esse canal, que terminou de processar a anterior.

```go
package main

import "fmt"

// requisicao carrega, além do valor, o canal pelo qual quem pediu quer
// receber a resposta. O serviço só precisa escrever nele, por isso chan<-.
type requisicao struct {
	valor    int
	resposta chan<- int
}

// servico atende uma requisição por vez e responde no canal que veio
// dentro da própria mensagem.
func servico(entrada <-chan requisicao) {
	for req := range entrada {
		req.resposta <- req.valor * 2
	}
}

func main() {
	entrada := make(chan requisicao)
	pronto := make(chan struct{})
	go func() {
		servico(entrada)
		close(pronto)
	}()

	for i := range 5 {
		resposta := make(chan int)
		entrada <- requisicao{valor: i, resposta: resposta}
		// Fica bloqueado até o serviço responder
		fmt.Println("resposta:", <-resposta)
	}

	// Sem mais requisições: o serviço termina
	close(entrada)
	<-pronto
}
```

## Parte 3 · Encerrando goroutines

Os geradores copiados nos exemplos da Parte 2 têm um defeito em comum: só terminam se alguém ler todos os valores. Esta parte trata de como mandar uma _goroutine_ parar, como saber que ela parou e o que acontece quando ninguém faz isso.

As duas palestras que mais aparecem aqui, a de Rob Pike (2012) e a de Sameer Ajmani (2013), são anteriores ao pacote `context`, que só entrou na biblioteca padrão no Go 1.7, em 2016. Foi o próprio Ajmani quem o apresentou, no [post](https://go.dev/blog/context) de julho de 2014. O que o `context` padronizou foi uma única técnica das palestras: o canal `quit`, fechado para avisar todo mundo de uma vez. É o `ctx.Done()`. O resto continua sem substituto, porque o contexto leva o sinal em um sentido só, de quem chama para quem é chamado, e nunca traz resultado de volta. O laço `for` com `select` e estado local, o canal de resposta que confirma a parada com um erro e o canal `nil` que desliga um `case` são escritos à mão hoje do mesmo jeito que em 2013. Esta parte mostra primeiro o canal `quit` das palestras e depois a forma com `context`, que é a que você vai encontrar em código de hoje.

### 🚏 Canal de parada (quit channel)

**Também conhecido como:** _quit channel_, canal `done`.

Um gerador sem fim, ou um consumidor que desiste no meio do caminho, deixa uma _goroutine_ bloqueada para sempre em um envio que ninguém vai receber. O canal de parada resolve isso. O gerador faz cada envio disputar, em um [`select`](#️-select-e-timeouts), com um canal `quit`. Quando quem consome não quer mais valores, fecha o `quit`, e o gerador termina em vez de ficar bloqueado. O padrão vem da palestra [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide), de Rob Pike.

No exemplo, o `contador` geraria números para sempre. A função principal lê os três primeiros e fecha o `quit`.

Repare que o gerador fecha a saída ao sair. A função principal lê a saída até ela ser fechada, e assim tem certeza de que o gerador terminou. A variante em que o gerador confirma a parada pelo próprio `quit` está em [parada com confirmação](#-parada-com-confirmação).

```go
package main

import "fmt"

// contador é um gerador sem fim: envia 0, 1, 2... até que o canal quit seja
// fechado. Ao sair, fecha o canal de saída.
func contador(quit <-chan struct{}) <-chan int {
	saida := make(chan int)
	go func() {
		defer close(saida)
		for i := 0; ; i++ {
			// O envio disputa com o sinal de parada: o que puder
			// prosseguir primeiro, vence.
			select {
			case saida <- i:
			case <-quit:
				return
			}
		}
	}()
	return saida
}

func main() {
	quit := make(chan struct{})
	valores := contador(quit)

	// Só queremos os três primeiros valores
	for range 3 {
		fmt.Println(<-valores)
	}

	// Manda o gerador parar. Sem isto ele ficaria bloqueado no próximo
	// envio para sempre.
	close(quit)

	// O gerador fecha a saída ao terminar: ler até o fechamento garante que
	// ele parou de fato.
	for range valores {
	}
	fmt.Println("o gerador parou")
}
```

### 🛑 Vazamento de goroutines e context

Uma _goroutine_ bloqueada em um canal que ninguém mais vai ler (ou escrever) nunca termina. Dizemos que ela vaza. O coletor de lixo não recolhe _goroutines_, então a memória e os recursos que ela segura ficam presos até o fim do programa. Em um programa curto isso passa despercebido. Em um servidor que roda por meses, é um vazamento de memória.

O gerador `sequenciaNumeros`, usado em vários exemplos, tem esse problema: ele só termina se alguém ler todos os valores. No [exemplo](./cancelamento/cancelamento.go), a função principal lê apenas os três primeiros e para. A _goroutine_ fica presa no envio do quarto valor. O programa mostra isso comparando `runtime.NumGoroutine()` antes e depois.

Esse contador é a forma rústica de achar um vazamento, e só funciona porque o programa é pequeno. Desde o Go 1.26, o coletor de lixo consegue apontar a _goroutine_ presa. O perfil `goroutineleak`, do pacote `runtime/pprof`, lista as _goroutines_ bloqueadas em um canal ou mutex que nenhuma _goroutine_ viva ainda alcança, e que por isso nunca vão acordar. É experimental: precisa de `GOEXPERIMENT=goroutineleakprofile` na compilação, e com ele o perfil aparece também em `/debug/pprof/goroutineleak`. Ele não pega tudo. Uma _goroutine_ presa em um canal que outra _goroutine_ viva ainda referencia não conta, porque em tese alguém ainda poderia ler.

A solução é a mesma do [canal de parada](#-canal-de-parada-quit-channel). Cada envio disputa, em um `select`, com um sinal de cancelamento. Só que, em vez de um canal `quit` próprio, o costume em Go é receber um `context.Context` e observar `ctx.Done()`, um canal que é fechado quando o contexto é cancelado. A vantagem é que o mesmo contexto atravessa várias funções e etapas de um _pipeline_, carrega prazos (`context.WithTimeout`) e cancela todo mundo de uma vez. Para ir mais fundo, veja o repositório sobre [context](https://github.com/cassiobotaro/contexto) e a segunda metade do artigo sobre [_pipelines_](https://go.dev/blog/pipelines).

Na versão cancelável, a função principal lê três valores e chama `cancel()`. Sem essa chamada, a _goroutine_ ficaria presa exatamente como a primeira.

```go
package main

import (
	"context"
	"fmt"
	"runtime"
)

// sequenciaNumeros é o gerador usado nos outros exemplos. Ele não é
// cancelável: se o consumidor parar de ler antes do fim, o envio bloqueia
// para sempre e a goroutine vaza.
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

// sequenciaNumerosCancelavel faz cada envio disputar com ctx.Done():
// se o contexto for cancelado, a goroutine desiste do envio e termina.
func sequenciaNumerosCancelavel(ctx context.Context, inicial, final int) <-chan int {
	saida := make(chan int)
	go func() {
		defer close(saida)
		for i := inicial; i <= final; i++ {
			select {
			case saida <- i:
			case <-ctx.Done():
				fmt.Println("gerador: cancelado, encerrando")
				return
			}
		}
	}()
	return saida
}

func main() {
	antes := runtime.NumGoroutine()

	// Sem cancelamento: lemos só os 3 primeiros valores e paramos.
	valores := sequenciaNumeros(1, 1000)
	for range 3 {
		fmt.Printf("valor: %v\n", <-valores)
	}
	// Ninguém mais vai ler de `valores`: a goroutine do gerador está presa
	// em `saida <- 4` e continuará assim até o programa terminar.
	fmt.Printf("goroutines presas: %d\n", runtime.NumGoroutine()-antes)

	// Com cancelamento: lemos os 3 primeiros valores e cancelamos.
	ctx, cancel := context.WithCancel(context.Background())
	cancelaveis := sequenciaNumerosCancelavel(ctx, 1, 1000)
	for range 3 {
		fmt.Printf("valor: %v\n", <-cancelaveis)
	}
	// Sem esta chamada a goroutine ficaria presa, como a anterior.
	cancel()
	// O gerador fecha o canal ao sair: drenar até o fechamento garante que
	// ele terminou de fato.
	for range cancelaveis {
	}
	fmt.Println("gerador cancelável encerrado")

	// Parada com confirmação (veja quit_confirmacao.go)
	quitComConfirmacao()
	// A mesma parada com context e errgroup (veja context_errgroup.go)
	paradaComErrgroup()
	// Vários sinais de parada combinados em um só (veja qualquer.go)
	combinarSinais()
}
```

### 🤝 Parada com confirmação

**Também conhecido como:** _shutdown_ com _ack_, _graceful stop_.

Mandar "pare" não garante que a _goroutine_ já parou. Se ela precisa liberar recursos antes de sair (fechar arquivos, encerrar conexões), quem pediu a parada deve esperar a confirmação. Na palestra [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide), Pike faz isso reaproveitando o próprio canal `quit`. Quem quer parar envia "pare", a _goroutine_ faz a limpeza e responde no mesmo canal. Por isso, [neste exemplo](./cancelamento/quit_confirmacao.go), o `quit` é um canal bidirecional, um dos raros casos em que isso é intencional.

A palestra [Advanced Go Concurrency Patterns](https://go.dev/talks/2013/advconc.slide), de Sameer Ajmani, chega ao mesmo resultado com [requisição e resposta](#-requisição-e-resposta). O método `Close` envia um canal de resposta por um `chan chan error` e espera nele. A _goroutine_ faz a limpeza e responde com o erro, se houver. O pedido desce e a resposta sobe pelo mesmo mecanismo.

O `context` faz só a metade de baixo. Ele leva o sinal de quem chama para quem é chamado e nunca traz nada de volta. Para a metade de cima, o costume hoje é juntar o `context` com um `errgroup`, o mesmo pacote visto na [variante do semáforo](#e-com-errgroup). A _goroutine_ roda dentro de `g.Go`, faz a limpeza ao ver `ctx.Done()` e devolve o erro. Quem quer parar chama `cancel()` e depois `g.Wait()`, que bloqueia até a _goroutine_ retornar e entrega esse erro. É o `Close` de Ajmani em duas chamadas, sem canal de resposta escrito à mão. A [segunda versão abaixo](./cancelamento/context_errgroup.go) faz isso, e a saída é a mesma da primeira: o gerador libera os recursos antes de o programa seguir.

Quando a confirmação precisa carregar mais do que um erro, o canal de resposta continua sendo a forma. E note que `context.WithCancelCause`, que existe desde o Go 1.20, deixa quem cancela dizer o motivo, lido do outro lado com `context.Cause(ctx)`. Isso é informação descendo, de quem cancela para quem é cancelado, e não substitui a confirmação.

```go
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
```

```go
package main

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

// tagarelaComContext envia mensagens até o contexto ser cancelado. Antes
// de sair faz a limpeza e devolve o motivo do cancelamento. A confirmação
// é o próprio retorno: quem chamou espera por ele com g.Wait.
func tagarelaComContext(ctx context.Context, nome string, saida chan<- string) error {
	for i := 0; ; i++ {
		select {
		case saida <- fmt.Sprintf("%s %d", nome, i):
		case <-ctx.Done():
			limpeza()
			return ctx.Err()
		}
	}
}

func paradaComErrgroup() {
	ctx, cancelar := context.WithCancel(context.Background())
	saida := make(chan string)

	var g errgroup.Group
	g.Go(func() error { return tagarelaComContext(ctx, "Bia", saida) })

	for range 3 {
		fmt.Println(<-saida)
	}
	// cancelar manda parar. Wait bloqueia até a goroutine retornar e traz
	// o erro dela: o pedido desce pelo contexto e a resposta sobe pelo
	// errgroup.
	cancelar()
	fmt.Println("gerador:", g.Wait())
}
```

### 🧩 Combinar sinais de parada (or-channel)

**Também conhecido como:** _or-channel_. Não confunda com o _or-done-channel_, do mesmo livro citado abaixo, que é outra técnica. Ele embrulha a leitura de um canal para que ela também respeite um sinal de parada.

Às vezes uma _goroutine_ deve parar quando _qualquer um_ de vários sinais chegar: o contexto da requisição, um sinal do sistema operacional, um prazo global. Em 2017, quando o livro citado abaixo saiu, cada um desses era um canal diferente, e a resposta era combinar os canais em um só. Hoje os três são contextos. O da requisição sempre foi, um _handler_ HTTP recebe `r.Context()`. O prazo é `context.WithTimeout`. E desde o Go 1.16, `signal.NotifyContext` devolve um contexto cancelado quando o sinal do sistema chega. Quando as origens são contextos, a resposta é derivar um do outro. O filho é cancelado quando o pai é, e a _goroutine_ fica com um único `case`, o `ctx.Done()`. A primeira metade do [exemplo](./cancelamento/qualquer.go) faz isso com as três origens.

O que sobra para combinar à mão é o que não é contexto: o canal `pronto` de outra _goroutine_, a saída de um gerador, qualquer `chan struct{}` fechado como sinal. A função `qualquer` combina esses canais em um só, que é fechado quando o primeiro deles fechar. Na segunda metade do exemplo, a _goroutine_ para quando o contexto acabar ou quando um colega terminar, o que vier primeiro.

A implementação usa uma _goroutine_ por canal de entrada, e a primeira a ser acordada fecha a saída. Ela toma dois cuidados. O primeiro é o `sync.Once`, que garante um único `close` mesmo que dois sinais cheguem juntos, já que fechar um canal duas vezes causa _panic_. O segundo é que cada _goroutine_ também observa a própria saída. Assim, quando um sinal vence, as demais terminam em vez de vazarem esperando canais que talvez nunca fechem.

Existem alternativas. Para duas ou três origens, um `select` explícito é o mais claro. Para o caso geral há a versão recursiva, que divide a lista ao meio, e o `reflect.Select`. As duas funcionam, mas são mais engenhosas do que claras, e os Go Proverbs lembram que "_Clear is better than clever_" e "_Reflection is never clear_". Para o sentido inverso, ligar um contexto a algo que não entende contextos, `context.AfterFunc(ctx, f)` roda `f` quando o contexto é cancelado, desde o Go 1.21.

De onde vem isso? Sinalizar a parada fechando um canal aparece no artigo sobre [_pipelines_](https://go.dev/blog/pipelines), com o canal `done`, e na palestra [Advanced Go Concurrency Patterns](https://go.dev/talks/2013/advconc.slide), de Sameer Ajmani (2013), cujo código fecha um canal `quit` para encerrar as _goroutines_ do `Merge`. Nenhum dos dois combina vários sinais em um só. O _or-channel_, com esse nome, é do livro _Concurrency in Go_, de Katherine Cox-Buday (O'Reilly, 2017, capítulo 4), que usa a versão recursiva.

```go
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
				// Outro sinal chegou primeiro: esta goroutine termina
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
	// goroutine fecha ao terminar; a goroutine é andaime, simula um colega
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
```

## Parte 4 · Controlando o ritmo

O que fazer quando quem produz e quem consome andam em velocidades diferentes. Cada padrão desta parte dá uma resposta: fazer o produtor esperar, limitar quantos executam ao mesmo tempo, limitar a taxa, agrupar o trabalho ou descartar o que ficou velho.

### 🚦 Contrapressão (backpressure)

**Também conhecido como:** _backpressure_, _bounded queue_ (fila limitada).

Contrapressão (_backpressure_) é o mecanismo pelo qual um consumidor lento faz o produtor diminuir o ritmo, em vez de deixar o trabalho se acumular sem limite. Aqui nada é descartado, e quem espera é o produtor. A resposta oposta é a da [janela deslizante](#-janela-deslizante), no fim desta parte, em que o produtor segue livre e os valores antigos são descartados.

Em Go esse mecanismo já vem embutido nos canais. A capacidade do canal é a folga máxima entre produtor e consumidor. Quando ela acaba, o envio bloqueia. O bloqueio propaga a lentidão do consumidor para trás, etapa por etapa, até chegar em quem gera os dados.

No exemplo, o produtor gera dez valores o mais rápido que consegue e o consumidor leva 200ms para processar cada um. A fila entre eles tem capacidade 3. Os primeiros valores entram de imediato, mas a partir do momento em que a fila enche, cada envio leva cerca de 200ms, que é justamente o ritmo do consumidor. O produtor não tem nenhum código para "esperar o consumidor", ele apenas escreve no canal.

Para tornar a espera visível, o produtor faz antes uma tentativa com `select` e `default`, que é a forma de perguntar "dá para enviar agora?" sem bloquear. Se não der, ele avisa que a fila está cheia e faz o envio bloqueante normal. É o mesmo mecanismo da nota abaixo sobre descarte de carga, mas aqui ele só observa e não descarta nada. Sem o `select` o comportamento seria o mesmo, só que em silêncio.

Repare em duas coisas que não acontecem. A memória não cresce, pois a fila tem um teto conhecido, e nenhum valor é perdido. O custo é que o produtor fica bloqueado, e isso precisa ser aceitável para quem está na ponta. Se quem produz é um _handler_ HTTP, por exemplo, bloquear pode significar segurar a conexão do cliente.

> **Quando bloquear não é opção.** Se o produtor não pode esperar, a alternativa é rejeitar o trabalho quando a fila está cheia, com um `select` e `default`. Se o envio não for possível de imediato, a chamada retorna um erro (um servidor devolveria algo como `503` ou `429`). Isso é descarte de carga (_load shedding_). A diferença para a janela deslizante é quem sai perdendo. Na janela é o valor mais antigo. No descarte de carga é o valor novo, que nem chega a entrar. Escolher entre bloquear, descartar o antigo ou rejeitar o novo depende do que o seu sistema pode tolerar. Para limitar a taxa ao longo do tempo, e não o tamanho da fila, veja o [sistema de ticket](#-sistema-de-ticket).

```go
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
		// select com default pergunta "dá para enviar agora?" sem bloquear.
		// Aqui ele serve apenas para observar a fila cheia: nada é descartado,
		// pois o default faz em seguida o envio bloqueante.
		select {
		case saida <- i:
			fmt.Printf("Produtor: enviou %d\n", i)
		default:
			fmt.Printf("Produtor: fila cheia, esperando para enviar %d\n", i)
			saida <- i
			fmt.Printf("Produtor: enviou %d após esperar\n", i)
		}
	}
}

// consumidorLento simula um trabalho que leva mais tempo do que a produção,
// como escrever em disco ou chamar um serviço externo.
func consumidorLento(entrada <-chan int, pronto chan<- struct{}) {
	// Sinaliza o término fechando o canal
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
```

### 🚥 Semáforo (paralelismo limitado)

**Também conhecido como:** _bounded parallelism_, limite de _goroutines_ em voo.

Um canal com buffer de capacidade `n` funciona como um semáforo. Enviar ocupa uma vaga, e bloqueia quando todas estão ocupadas. Receber libera uma vaga. Com isso dá para limitar quantas _goroutines_ executam um trecho ao mesmo tempo sem criar um grupo fixo. Cada tarefa tem sua própria _goroutine_, mas só `n` avançam de cada vez. A técnica aparece com o nome de _bounded parallelism_ no artigo sobre [_pipelines_](https://go.dev/blog/pipelines).

Qual a diferença para os padrões vizinhos? O [grupo de trabalhadores](#-grupo-de-trabalhadores-pool-of-workers) fixa o número de _goroutines_. O [sistema de ticket](#-sistema-de-ticket) limita a taxa ao longo do tempo. O semáforo limita quantas tarefas executam ao mesmo tempo. Este é um dos poucos casos em que o buffer do canal não é um ajuste fino, porque a capacidade do canal é o próprio limite.

No exemplo, dez tarefas são disparadas de uma vez, mas o semáforo tem três vagas. A saída mostra que o número de tarefas ativas nunca passa de três. O contador atômico (`sync/atomic`) serve apenas para observar isso e não faz parte do padrão. O `sync.WaitGroup` aguarda o término de todas as tarefas.

```go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// executarTarefas dispara uma goroutine por tarefa, mas deixa no máximo
// `limite` delas executarem ao mesmo tempo.
func executarTarefas(tarefas, limite int) {
	// Um canal com buffer funciona como semáforo: cada valor no buffer é uma
	// vaga ocupada. Enviar bloqueia quando as `limite` vagas estão ocupadas.
	sem := make(chan struct{}, limite)

	var wg sync.WaitGroup
	// Contador usado apenas para observar quantas tarefas estão ativas;
	// ele não faz parte do padrão.
	var ativas atomic.Int32

	// Cada tarefa tem sua própria goroutine, mas só `limite` avançam por vez
	for i := range tarefas {
		wg.Go(func() {
			sem <- struct{}{}        // ocupa uma vaga (bloqueia se não houver)
			defer func() { <-sem }() // libera a vaga ao terminar

			fmt.Printf("tarefa %2d começou, ativas: %d\n", i+1, ativas.Add(1))
			time.Sleep(100 * time.Millisecond)
			ativas.Add(-1)
		})
	}

	wg.Wait()
}

func main() {
	// Dez tarefas, no máximo três ao mesmo tempo
	executarTarefas(10, 3)

	// O mesmo com errgroup (veja com_errgroup.go): a tarefa 2 falha e as
	// que ainda não começaram são canceladas
	if err := executarTarefasErrgroup(10, 3, 2); err != nil {
		fmt.Println("errgroup:", err)
	}
}
```

#### E com errgroup?

O canal de vagas e o `WaitGroup` fazem duas coisas que o padrão sempre precisa: limitar e esperar. O pacote [`golang.org/x/sync/errgroup`](https://pkg.go.dev/golang.org/x/sync/errgroup) junta as duas em um tipo só e acrescenta a terceira, que os exemplos até aqui ignoraram: o erro. `g.SetLimit(3)` é o canal de três vagas. `g.Go` dispara a tarefa, e bloqueia quando as vagas acabam. `g.Wait` espera todas, como o `WaitGroup`, e devolve o primeiro erro que alguma tarefa retornou. Com `errgroup.WithContext`, esse primeiro erro cancela um contexto, e as tarefas que ainda não começaram podem desistir olhando `ctx.Err()`.

A [versão abaixo](./semaforo/com_errgroup.go) faz isso. Com três vagas e a tarefa 2 falhando, as tarefas 1 a 3 começam, e as outras sete são canceladas antes de começar, porque quando elas conseguem uma vaga o contexto já foi cancelado. A tarefa que falha é andaime, existe só para mostrar o cancelamento.

O custo é uma dependência fora da biblioteca padrão. É a única deste repositório. Vale a pena quando as tarefas devolvem erro e uma falha deve interromper as demais, que é o caso comum em código de produção. Quando as tarefas não falham, ou quando cada erro deve ser tratado por conta própria, o canal com buffer e o `WaitGroup` bastam.

```go
package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
)

// executarTarefasErrgroup faz o mesmo que executarTarefas, com um
// errgroup.Group no lugar do canal e do WaitGroup: SetLimit é o número de
// vagas, Go bloqueia quando elas acabam e Wait espera todas e devolve o
// primeiro erro. A tarefa de número `falha` devolve erro para mostrar o
// que acontece com as outras; ela é andaime, não faz parte do padrão.
func executarTarefasErrgroup(tarefas, limite, falha int) error {
	// O contexto é cancelado no primeiro erro
	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(limite)

	// Contador usado apenas para observar quantas tarefas estão ativas;
	// ele não faz parte do padrão.
	var ativas atomic.Int32

	for i := range tarefas {
		g.Go(func() error {
			// As tarefas que ainda não começaram desistem
			if ctx.Err() != nil {
				fmt.Printf("errgroup: tarefa %2d cancelada\n", i+1)
				return ctx.Err()
			}

			fmt.Printf("errgroup: tarefa %2d começou, ativas: %d\n", i+1, ativas.Add(1))
			time.Sleep(100 * time.Millisecond)
			ativas.Add(-1)

			if i+1 == falha {
				return fmt.Errorf("tarefa %d falhou", i+1)
			}
			return nil
		})
	}

	return g.Wait()
}
```

### 🎫 Sistema de ticket

**Também conhecido como:** _rate limiting_, _throttling_. São nomes aproximados, porque aqui a taxa é fixa, sem o saldo para rajadas de um _token bucket_ (veja a nota sobre rajada abaixo).

Um sistema de ticket controla quando um trabalho pode ser executado. Serve para limitar o uso de um recurso ao longo de um período, como uma API que aceita 15 chamadas a cada 15 minutos.

No exemplo, a bilheteria emite no máximo 10 tickets por segundo. A função principal envia 31 trabalhos pelo canal, e eles saem a 10 por segundo, levando cerca de três segundos no total.

O ticket limita a taxa ao longo do tempo. Para limitar quantas tarefas executam ao mesmo tempo, veja o [semáforo](#-semáforo-paralelismo-limitado).

O trabalhador pega um trabalho e fica bloqueado até receber um ticket. A ordem importa. Como o trabalho é lido primeiro, o trabalhador encerra sem gastar um ticket à toa quando o canal de trabalhos é fechado.

> **Nota sobre rajada (_burst_).** Esta implementação emite um ticket a cada `timeout/nTickets`, garantindo o teto mesmo se o consumidor for mais lento do que o ticker. Em troca, ela não permite rajadas, pois não há um saldo inicial de `nTickets` para ser consumido de uma só vez. Se você precisar de _rate limiting_ com tolerância a rajadas (_token bucket_, isto é, rajada de até N seguida de reposição a `T/N`), use [`golang.org/x/time/rate`](https://pkg.go.dev/golang.org/x/time/rate).

```go
package main

import (
	"context"
	"fmt"
	"time"
)

type (
	Trabalho func()
	ticket   int
)

func trabalhador(tickets <-chan ticket, trabalhos <-chan Trabalho) {
	for {
		// Lê o trabalho primeiro: se o canal foi fechado, encerra
		// sem gastar um ticket.
		trabalho, ok := <-trabalhos
		if !ok {
			return // canal de trabalhos fechado
		}
		<-tickets  // espera autorização antes de executar
		trabalho() // executa um trabalho
	}
}

// bilheteria emite, no máximo, nTickets por intervalo `timeout`,
// ou seja, um ticket a cada `timeout/nTickets`. Garante o teto mesmo com consumidor
// lento, em troca de não permitir rajadas (nenhuma janela "extra" no início).
func bilheteria(ctx context.Context, tickets chan<- ticket, timeout time.Duration, nTickets int) {
	intervalo := timeout / time.Duration(nTickets)
	ticker := time.NewTicker(intervalo)
	defer ticker.Stop()

	var i int
	for {
		select {
		case tickets <- ticket(i):
			i++
		case <-ctx.Done():
			return
		}

		// espera o intervalo mínimo antes de emitir o próximo ticket
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	tickets := make(chan ticket)
	trabalhos := make(chan Trabalho)
	pronto := make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go bilheteria(ctx, tickets, 1*time.Second, 10)
	go func() {
		trabalhador(tickets, trabalhos)
		// Sinaliza o término fechando o canal
		close(pronto)
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

### 📦 Processamento em lote (batch processing)

**Também conhecido como:** _batching_, _micro-batching_.

Um processamento em lote (_batch processing_) agrupa itens que chegam um por um, para que o consumidor os processe em blocos. Um canal de descarga força o envio do lote antes de ele encher, e um canal de conclusão avisa quando o último lote foi processado.

Um exemplo: em vez de gravar cada item no banco assim que ele chega, o programa junta 100 itens, ou 100ms de itens, e grava tudo em uma requisição só.

No exemplo, o lote tem capacidade para três itens. Quando o terceiro chega, o lote enche e segue para o canal de saída.

Um lote pode ser descarregado de três formas:

- Quando ele enche.
- Quando o `intervalo` passa sem que ele tenha enchido. Isso é feito com um `time.Ticker` dentro do `select`, e evita que um item fique esperando companhia por tempo indeterminado.
- Sob demanda, pelo canal `descarga`.

No exemplo, o item 6 é enviado sozinho e sai pelo intervalo de 100ms.

Repare que, depois de enviar um lote, o código cria um novo _slice_ em vez de reaproveitar o anterior com `buf[:0]`. O consumidor pode ainda estar lendo o lote enviado, e reutilizar o mesmo _array_ de apoio sobrescreveria dados em uso. Parece uma otimização óbvia, mas quebraria o programa.

Se a entrada for fechada com itens no buffer, o lote parcial ainda é enviado.

```go
package main

import (
	"fmt"
	"time"
)

type req struct {
	valor int
}

func processar(lote []req) {
	fmt.Println("processando lote com valores: ", lote)
}

func processadorLotes(entrada <-chan []req) <-chan struct{} {
	pronto := make(chan struct{})
	go func() {
		for lote := range entrada {
			processar(lote)
		}
		// Sinaliza o término do processamento fechando o canal
		close(pronto)
	}()
	return pronto
}

// processamentoLotes agrupa os itens da entrada em lotes. Um lote é enviado
// quando enche (`tamanhoLote`), quando passa o `intervalo` sem que ele tenha
// enchido, ou quando chega um sinal manual pelo canal `descarga`.
func processamentoLotes(entrada <-chan req, descarga <-chan struct{}, tamanhoLote int, intervalo time.Duration) <-chan []req {
	saida := make(chan []req)
	go func() {
		defer close(saida)
		buf := make([]req, 0, tamanhoLote)

		ticker := time.NewTicker(intervalo)
		defer ticker.Stop()

		// descarregar envia o lote atual, se houver algo nele
		descarregar := func() {
			if len(buf) == 0 {
				return
			}
			saida <- buf
			// Um novo slice é criado em vez de reaproveitar com buf[:0]:
			// o consumidor pode ainda estar lendo o lote enviado, e reutilizar
			// o mesmo array de apoio sobrescreveria dados em uso (aliasing).
			buf = make([]req, 0, tamanhoLote)
		}

		for {
			select {
			// enquanto houver itens para processar
			case item, ok := <-entrada:
				if !ok {
					// envia o que tiver no buffer antes de sair
					descarregar()
					// para o loop quando o canal de entrada for fechado
					return
				}
				// Adiciona o item no buffer
				buf = append(buf, item)
				// se o buffer estiver cheio, descarrega
				if len(buf) == tamanhoLote {
					descarregar()
				}

			// Se o intervalo passou, descarrega o que tiver no buffer
			case <-ticker.C:
				descarregar()

			// Se receber um sinal de descarga, descarrega o que tiver no buffer
			case <-descarga:
				descarregar()
			}
		}
	}()
	return saida
}

func main() {
	entrada := make(chan req)
	descarga := make(chan struct{})

	// inicia de forma concorrente o processamento em lotes:
	// lotes de 3 itens ou 100ms, o que acontecer primeiro
	saida := processamentoLotes(entrada, descarga, 3, 100*time.Millisecond)
	// O consumidor de lotes será iniciado de forma concorrente
	pronto := processadorLotes(saida)

	entrada <- req{valor: 1}
	entrada <- req{valor: 2}
	entrada <- req{valor: 3}

	// Envia mais dois itens e força a descarga do lote
	// pelo canal de descarga
	entrada <- req{valor: 4}
	entrada <- req{valor: 5}
	descarga <- struct{}{}

	// Envia um item e espera: o lote não enche, mas o intervalo
	// de 100ms passa e ele é descarregado mesmo assim
	entrada <- req{valor: 6}
	time.Sleep(150 * time.Millisecond)

	// Envia mais dois itens, não o suficiente para descarregar
	// o lote.
	entrada <- req{valor: 7}
	entrada <- req{valor: 8}
	// Eles serão processados mesmo assim.

	close(entrada)

	// Aguarda todo o processamento do processador de lotes
	// antes de encerrar o programa
	<-pronto
}
```

### 🪟 Janela deslizante

**Também conhecido como:** _drop-oldest buffer_. Evite tratar _ring buffer_ como sinônimo. O _ring buffer_ é uma forma de armazenar os dados, enquanto a janela deslizante é a regra de descarte, em que sai sempre o mais antigo.

Uma janela deslizante (_sliding window_) impede que um leitor lento trave um escritor rápido. É a resposta oposta à da [contrapressão](#-contrapressão-backpressure): em vez de o produtor esperar, os valores mais velhos são descartados. A ordem das entregas é mantida, mas um consumidor lento perde os valores que já saíram da janela.

No exemplo, o produtor envia um valor por segundo e o consumidor leva quatro segundos para processar cada um. A janela guarda três valores, então, à medida que ela desliza, os mais antigos são descartados.

Para fazer a janela deslizante, uma única _goroutine_ é dona de todo o estado (uma fila com tamanho máximo fixo) e usa um `select` para reagir ao que acontecer primeiro. Se chega um valor da entrada, ele entra na fila, e o mais antigo é descartado caso ela esteja cheia. Se o consumidor está pronto para receber, o primeiro da fila é enviado.

O truque aqui é o canal `nil`, visto em [select](#️-select-e-timeouts). Como um `case` cujo canal é `nil` nunca é escolhido, dá para ligar e desligar cada `case` conforme o estado da fila. Quando a fila está vazia, o canal de envio fica `nil` e o `case` de envio é desabilitado, pois não há o que enviar. Quando a entrada é fechada, a variável `entrada` passa a valer `nil` e o `case` de recebimento é desabilitado. Daí em diante só resta esvaziar a fila.

Como só uma _goroutine_ toca a fila, não há disputa entre produtor e consumidor pelo estado. É a técnica da [goroutine dona do estado](#-goroutine-dona-do-estado). Uma versão anterior deste exemplo usava um canal com buffer compartilhado por duas _goroutines_ e tinha uma corrida sutil que podia travar o programa.

```go
package main

import (
	"fmt"
	"time"
)

// janelaDeslizante mantém apenas os `tamanho` itens mais recentes vindos de
// `entrada`, descartando o mais antigo quando a janela enche. Uma única
// goroutine é dona de todo o estado (a fila), então não há disputa entre
// produtor e consumidor pelo buffer.
func janelaDeslizante(entrada <-chan int, saida chan<- int, tamanho int) {
	defer close(saida)
	var fila []int

	for entrada != nil || len(fila) > 0 {
		// O case de envio só é habilitado quando há algo na fila:
		// um canal nil nunca é selecionado, o que desabilita o case.
		var envio chan<- int
		var cabeca int
		if len(fila) > 0 {
			envio = saida
			cabeca = fila[0]
		}

		select {
		case valor, ok := <-entrada:
			if !ok {
				// Entrada fechada: desabilita este case (canal nil)
				// e continua apenas drenando a fila.
				entrada = nil
				continue
			}
			if len(fila) == tamanho {
				// Janela cheia, descarta o mais antigo e adiciona o novo
				fmt.Printf("Janela Deslizante: Buffer cheio, descartou %v para adicionar %v.\n", fila[0], valor)
				// fila[1:] não libera memória na hora: o array de apoio é
				// mantido até o próximo append realocar. Para uma janela
				// pequena isso é irrelevante, mas é bom saber.
				fila = fila[1:]
			}
			fila = append(fila, valor)

		case envio <- cabeca:
			fmt.Printf("Janela Deslizante: Enviou %v para o consumidor.\n", cabeca)
			fila = fila[1:]
		}
	}
}

// sequenciaNumeros, aqui, avisa a cada envio e faz uma pausa de um segundo
// entre eles, para que o produtor seja mais rápido do que o consumidor.
func sequenciaNumeros(inicial, final int) <-chan int {
	saida := make(chan int)
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

func leitorLento(entrada <-chan int, pronto chan<- struct{}) {
	for valor := range entrada {
		fmt.Printf("Consumidor: Recebeu %v\n", valor)
		time.Sleep(4 * time.Second)
	}
	// Fechar o canal é o idioma para sinalizar um evento único
	close(pronto)
}

func main() {
	valores := sequenciaNumeros(1, 10)
	saida := make(chan int)
	pronto := make(chan struct{})
	go leitorLento(saida, pronto)
	janelaDeslizante(valores, saida, 3)
	<-pronto
	fmt.Println("Fim da execução.")
}
```

## Parte 5 · Padrões avançados

Padrões que combinam várias peças das partes anteriores, como canais de resposta, `select`, buffer, timeout e `context`. Vale ler as outras partes antes.

### 🔐 Goroutine dona do estado

**Também conhecido como:** monitor, confinamento, ator. O último é aproximado. No modelo de atores a mensagem vai para o ator pelo nome, e aqui ela vai por canais. É a mesma diferença entre Erlang e Go comentada na introdução.

O provérbio diz "_Don't communicate by sharing memory, share memory by communicating_", ou seja, não comunique compartilhando memória, compartilhe memória comunicando. Em vez de proteger uma variável com mutex e deixar várias _goroutines_ mexerem nela, uma única _goroutine_ é dona do estado, e as outras pedem alterações e leituras por canais. Não há corrida porque só uma _goroutine_ toca o dado. A palestra [Advanced Go Concurrency Patterns](https://go.dev/talks/2013/advconc.slide), de Sameer Ajmani (2013), apresenta a técnica como um laço `for` com `select` e estado local, e a resume assim: a _goroutine_ serializa o acesso ao próprio estado mutável, sem mutex, sem variável de condição e sem _callback_. É a primeira das três técnicas da palestra. As outras duas, o canal de resposta e o canal `nil`, estão em [parada com confirmação](#-parada-com-confirmação) e em [select](#️-select-e-timeouts). A [janela deslizante](#-janela-deslizante) já usa essa técnica por dentro. Aqui ela é o assunto principal.

No exemplo, a _goroutine_ `contador` é dona de um mapa de contagem por chave. Três _goroutines_ enviam mil incrementos cada uma pelo canal `incrementar`, e as leituras usam o canal `consultar`, com o canal de resposta dentro da mensagem, como em [requisição e resposta](#-requisição-e-resposta). O `select` atende um pedido por vez. Para encerrar, a função principal fecha `incrementar`.

```go
package main

import (
	"fmt"
	"sync"
)

// consulta pede a contagem de uma chave e carrega o canal de resposta,
// como no exemplo de requisição e resposta.
type consulta struct {
	chave    string
	resposta chan<- int
}

// contador é a goroutine dona do estado: só ela toca o mapa `contagem`.
// As demais goroutines pedem alterações e leituras pelos canais.
// Termina quando o canal `incrementar` é fechado.
func contador(incrementar <-chan string, consultar <-chan consulta) {
	contagem := make(map[string]int)
	for {
		select {
		case chave, ok := <-incrementar:
			if !ok {
				return
			}
			contagem[chave]++
		case c := <-consultar:
			c.resposta <- contagem[c.chave]
		}
	}
}

func main() {
	incrementar := make(chan string)
	consultar := make(chan consulta)
	pronto := make(chan struct{})
	go func() {
		contador(incrementar, consultar)
		close(pronto)
	}()

	// Várias goroutines incrementam ao mesmo tempo, sem mutex:
	// os pedidos são atendidos um por vez pela goroutine dona.
	var wg sync.WaitGroup
	for _, chave := range []string{"gopher", "gopher", "marmota"} {
		wg.Go(func() {
			for range 1000 {
				incrementar <- chave
			}
		})
	}
	wg.Wait()

	for _, chave := range []string{"gopher", "marmota"} {
		resposta := make(chan int)
		consultar <- consulta{chave: chave, resposta: resposta}
		fmt.Printf("dona do estado: %s = %d\n", chave, <-resposta)
	}

	// Sem mais incrementos: a goroutine dona termina
	close(incrementar)
	<-pronto

	// A mesma contagem feita com mutex (veja com_mutex.go)
	comMutex()
}
```

#### E com mutex?

O contraponto também vem de Pike, no provérbio "_Channels orchestrate; mutexes serialize_". Se tudo o que você precisa é serializar o acesso a um contador ou a um mapa, um `sync.Mutex` é mais simples e mais claro. A [versão abaixo](./dono_do_estado/com_mutex.go) faz isso e produz o mesmo resultado.

Quando a _goroutine_ dona do estado compensa?

- Quando há regras sobre _como_ o estado muda, como validação, ordem ou eventos.
- Quando ela precisa reagir a vários canais com `select`, como entradas, prazos e cancelamento. É o caso da janela deslizante.
- Quando o estado tem ciclo de vida próprio.

Se nada disso se aplica, use o mutex.

```go
package main

import (
	"fmt"
	"sync"
)

// contadorMutex resolve o mesmo problema serializando o acesso ao mapa.
// Quando tudo o que se precisa é proteger um dado, esta versão é mais
// simples e mais clara do que uma goroutine dona do estado.
type contadorMutex struct {
	mu       sync.Mutex
	contagem map[string]int
}

func (c *contadorMutex) incrementar(chave string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.contagem[chave]++
}

func (c *contadorMutex) consultar(chave string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.contagem[chave]
}

func comMutex() {
	c := contadorMutex{contagem: make(map[string]int)}

	var wg sync.WaitGroup
	for _, chave := range []string{"gopher", "gopher", "marmota"} {
		wg.Go(func() {
			for range 1000 {
				c.incrementar(chave)
			}
		})
	}
	wg.Wait()

	for _, chave := range []string{"gopher", "marmota"} {
		fmt.Printf("mutex: %s = %d\n", chave, c.consultar(chave))
	}
}
```

### 🏁 Primeiro a responder

**Também conhecido como:** _hedged request_, réplicas, `First` (o nome da função na palestra de Pike).

Para não depender do servidor mais lento, envie a mesma requisição a várias réplicas e use a primeira resposta que chegar. É a técnica que Rob Pike usa no exemplo da busca do Google, na palestra [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide), para reduzir a latência de cauda. Combinada com o timeout visto em [select](#️-select-e-timeouts), o resultado é o que Pike descreve como um programa rápido, replicado e robusto.

A palestra é de 2012, e o `First` de Pike só lê a primeira resposta. As perdedoras continuam trabalhando até o fim e jogam o resultado fora. Em uma chamada de rede, isso é uma requisição a mais no servidor por réplica. No [exemplo](./primeiro/primeiro.go), `primeiro` recebe um `context.Context`, deriva um filho com `context.WithCancel`, passa esse filho a cada réplica e cancela ao retornar. As perdedoras desistem no próximo `ctx.Done()`. Uma réplica de verdade faria o mesmo com `http.NewRequestWithContext`, que interrompe a requisição inteira quando o contexto é cancelado. É o [cancelamento](#-vazamento-de-goroutines-e-context) da Parte 3 aplicado ao padrão.

Repare no canal com buffer de tamanho `len(replicas)`. Mesmo com o cancelamento, uma perdedora pode terminar entre a chegada da vencedora e o `cancel()`. Com um canal sem buffer ela ficaria bloqueada no envio para sempre, pois ninguém mais vai ler. Com uma vaga por réplica, ela deposita a resposta e termina.

No exemplo, as réplicas são simuladas com uma espera aleatória de até 100ms, então a vencedora muda a cada execução. A segunda parte combina `primeiro` com um prazo de 20ms. O prazo vem em um `context.WithTimeout`, e `primeiro` o repassa às réplicas, então o mesmo `ctx.Done()` que encerra a espera encerra também as réplicas.

```go
package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"
)

// replica simula um servidor cuja latência varia a cada chamada. Se o
// contexto for cancelado antes da resposta, ela desiste, como faria uma
// requisição HTTP feita com http.NewRequestWithContext.
func replica(nome string) func(context.Context, string) (string, error) {
	return func(ctx context.Context, consulta string) (string, error) {
		select {
		case <-time.After(rand.N(100 * time.Millisecond)):
			return fmt.Sprintf("%s respondeu a %q", nome, consulta), nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}

// primeiro envia a mesma consulta a todas as réplicas e devolve a primeira
// resposta que chegar. Ao retornar, cancela o contexto das outras, para
// que as perdedoras parem de trabalhar em vez de responder à toa.
func primeiro(ctx context.Context, consulta string, replicas ...func(context.Context, string) (string, error)) (string, error) {
	ctx, cancelar := context.WithCancel(ctx)
	defer cancelar()

	// O buffer tem uma vaga por réplica. Uma perdedora pode terminar entre
	// a chegada da vencedora e o cancel, e sem a vaga ficaria presa no
	// envio para sempre, pois ninguém mais vai ler do canal.
	respostas := make(chan string, len(replicas))
	for _, r := range replicas {
		go func() {
			if resposta, err := r(ctx, consulta); err == nil {
				respostas <- resposta
			}
		}()
	}

	select {
	case resposta := <-respostas:
		return resposta, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func main() {
	replicas := []func(context.Context, string) (string, error){
		replica("réplica 1"),
		replica("réplica 2"),
		replica("réplica 3"),
	}

	resposta, err := primeiro(context.Background(), "golang", replicas...)
	if err != nil {
		fmt.Println("erro:", err)
		return
	}
	fmt.Println(resposta)

	// Combinado com timeout: usa a resposta mais rápida, desde que chegue
	// em até 20ms. O prazo vem no contexto, e primeiro o repassa às réplicas.
	ctx, cancelar := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancelar()
	resposta, err = primeiro(ctx, "csp", replicas...)
	if err != nil {
		fmt.Println("tempo esgotado: nenhuma réplica respondeu em 20ms")
		return
	}
	fmt.Println(resposta)
}
```

### 💓 Heartbeat

**Também conhecido como:** sinal de vida, _liveness_.

Um trabalhador que roda por muito tempo pode travar sem que ninguém perceba. Com um _heartbeat_ (batimento), ele emite um sinal em um canal a cada intervalo, e o supervisor usa `select` com timeout para decidir que o trabalhador morreu se o sinal não chegar. Com isso dá para diferenciar um trabalhador que está demorando de um que parou de responder. A forma apresentada aqui segue a do livro _Concurrency in Go_, de Katherine Cox-Buday (O'Reilly, 2017).

Repare em dois detalhes do exemplo. O primeiro é que o batimento é enviado com `select` e `default`. Se ninguém estiver ouvindo, o sinal se perde, e o trabalho nunca fica bloqueado por causa dele. O segundo é que o timeout do supervisor usa `time.After` dentro do laço, como no timeout por mensagem visto em [select](#️-select-e-timeouts). Assim, qualquer batimento ou resultado renova o prazo.

No exemplo, o trabalhador leva três intervalos e meio para produzir cada resultado e, de propósito, trava ao produzir o terceiro. O supervisor fica dois intervalos sem notícia e o declara morto. Ao sair, o supervisor cancela o contexto, para que o trabalhador termine caso volte a responder. Esperar por ele não faria sentido, já que um trabalhador travado de verdade pode nunca voltar.

```go
package main

import (
	"context"
	"fmt"
	"time"
)

// trabalhador produz um resultado a cada 3 intervalos e meio e, enquanto
// isso, emite um batimento a cada intervalo para mostrar que continua vivo.
// No terceiro resultado ele trava por `travamento`, e os batimentos param.
func trabalhador(ctx context.Context, intervalo, travamento time.Duration) (<-chan struct{}, <-chan int) {
	batimento := make(chan struct{})
	resultados := make(chan int)
	go func() {
		defer close(resultados)
		pulso := time.NewTicker(intervalo)
		defer pulso.Stop()
		trabalho := time.NewTicker(3*intervalo + intervalo/2)
		defer trabalho.Stop()

		for i := 1; ; {
			select {
			case <-ctx.Done():
				return
			case <-pulso.C:
				// Envio não bloqueante: se ninguém estiver ouvindo,
				// o batimento é perdido e o trabalho segue.
				select {
				case batimento <- struct{}{}:
				default:
				}
			case <-trabalho.C:
				if i == 3 {
					// Simula um travamento
					time.Sleep(travamento)
				}
				select {
				case resultados <- i:
					i++
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return batimento, resultados
}

// supervisor acompanha o trabalhador e o declara morto se ficar dois
// intervalos sem batimento e sem resultado.
func supervisor(intervalo, travamento time.Duration) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	batimento, resultados := trabalhador(ctx, intervalo, travamento)

	for {
		select {
		case <-batimento:
			fmt.Println("batimento")
		case r := <-resultados:
			fmt.Println("resultado:", r)
		case <-time.After(2 * intervalo):
			fmt.Println("trabalhador não responde")
			return
		}
	}
}

func main() {
	supervisor(100*time.Millisecond, 1*time.Second)
}
```

## Curiosidade

Um exemplo que não resolve nenhum problema do dia a dia, mas mostra o quanto uma _goroutine_ é barata.

### ⛓️ Daisy-chain

**Também conhecido como:** corrente de _goroutines_, telefone sem fio.

_Goroutines_ são baratas, e é comum ter dezenas de milhares delas. Este exemplo, tirado da palestra [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide), liga 10 mil _goroutines_ em uma corrente, cada uma somando 1 ao valor que recebe da vizinha da direita e passando o resultado para a esquerda. O valor 1 entra por uma ponta e sai 10001 pela outra.

Ninguém usa isso no dia a dia. O exemplo serve para mostrar que dividir o trabalho em pedaços bem pequenos não custa caro. Criar 10 mil _threads_ do sistema operacional para somar 1 seria impensável. Com _goroutines_, o programa termina em uma fração de segundo.

```go
package main

import "fmt"

// elo recebe um valor da vizinha da direita, soma 1 e passa para a esquerda.
func elo(esquerda chan<- int, direita <-chan int) {
	esquerda <- 1 + <-direita
}

func main() {
	const n = 10000

	// Monta a corrente da esquerda para a direita: cada goroutine fica
	// bloqueada esperando o valor da vizinha.
	pontaEsquerda := make(chan int)
	esquerda := pontaEsquerda
	var direita chan int
	for range n {
		direita = make(chan int)
		go elo(esquerda, direita)
		esquerda = direita
	}

	// Solta o primeiro valor na ponta direita...
	go func() { direita <- 1 }()
	// ...e espera ele atravessar as 10 mil goroutines.
	fmt.Println(<-pontaEsquerda)
}
```
