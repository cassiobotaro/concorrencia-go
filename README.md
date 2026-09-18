# Concorrência em Go

Go é fundamentada no modelo CSP (Communicating sequential processes) proposto por Tony Hoare. Neste modelo, os dados são compartilhados enviando mensagens através de canais.

Há uma nuance histórica aqui. No CSP original de Hoare, um processo envia mensagens diretamente para outro processo, identificado pelo nome; é o caminho seguido por Erlang. Go vem de outro ramo da família, o das linguagens Newsqueak, Alef e Limbo, em que o canal é um valor de primeira classe: pode ser guardado em variáveis, passado como parâmetro e até enviado por outro canal. Os dois modelos são equivalentes, mas se expressam de forma diferente. A analogia de Rob Pike é a de escrever em um arquivo pelo nome (processo, Erlang) ou por meio de um descritor de arquivo (canal, Go).

Outra ideia que acompanha todo o texto: [concorrência não é paralelismo](https://go.dev/blog/waza-talk). Concorrência é a composição de computações que executam de forma independente, ou seja, uma maneira de estruturar o programa. Paralelismo é executar várias computações ao mesmo tempo. Um programa concorrente pode rodar em um único processador, e um programa bem estruturado para concorrência tende a paralelizar bem quando há mais processadores disponíveis.

As explicações e exemplos são altamente inspirados na [apresentação](https://github.com/andrebq/andrebq.github.io) do @andrebq.

Outras influências:

- O [artigo](https://go.dev/blog/pipelines) sobre _pipelines_ e cancelamento em Go.
- A palestra [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide) de Rob Pike (Google I/O 2012), de onde vêm os geradores, o fan-in, os timeouts com `select` e o canal de parada.
- Os [Go Proverbs](https://go-proverbs.github.io/), também de Rob Pike (Gopherfest 2015): "_Don't communicate by sharing memory, share memory by communicating_", "_Concurrency is not parallelism_", "_Channels orchestrate; mutexes serialize_" e "_Clear is better than clever_".

Aqui serão apresentados alguns padrões de concorrência, porém sugiro também a leitura sobre [context](https://github.com/cassiobotaro/contexto), [select](https://gobyexample.com/select), [canais com buffer](https://gobyexample.com/channel-buffering) e outros mecanismos de controle de concorrência.

Um aviso sobre nomes: os mesmos padrões aparecem com nomes diferentes em livros, artigos e outras linguagens, por isso cada seção traz uma linha "Também conhecido como". "Produtor" e "consumidor" são papéis, não padrões: quase todo exemplo tem os dois. Eles aparecem como nomes alternativos de [Geradores](#-geradores) e [Trabalhador](#-trabalhador-worker) porque são as seções em que esses papéis estão isolados.

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

**Também conhecido como:** produtor, _source_. É o mesmo papel do `produtor` da seção de [contrapressão](#-contrapressão-backpressure).

Geradores são funções que iniciam uma _goroutine_ para escrever uma lista de valores em um canal que é retornado para quem acionou a função.

No exemplo, uma sequência de números inteiros é gerada e enviada para um canal.

A função principal (_main_) irá realizar a leitura do canal e imprimir os valores. Essa é uma característica interessante sobre canais, quando utilizados com o _range_, a iteração continuará até que o canal seja fechado.

> **Atenção:** este gerador não é cancelável: se o consumidor parar de ler antes do fim, a _goroutine_ vaza. Veja [Cancelamento](#-cancelamento-e-vazamento-de-goroutines).

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

## 🎛️ Select, timeout e quit channel

O `select` é uma estrutura de controle exclusiva para concorrência: parece um `switch`, mas cada `case` é uma comunicação (um envio ou um recebimento em um canal). Ele bloqueia até que alguma das comunicações possa prosseguir; se várias puderem ao mesmo tempo, escolhe uma de forma pseudoaleatória; e, se houver um `default`, não bloqueia. Rob Pike diz na palestra [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide) que o `select` é a razão de canais e _goroutines_ serem embutidos na linguagem, e não uma biblioteca.

Vários exemplos mais adiante dependem dele. Aqui vemos três usos básicos, todos com o mesmo gerador `tagarela`, que fala cada vez mais devagar:

- **Timeout por mensagem.** `time.After` devolve um canal que recebe um valor depois do tempo indicado. Colocado em um `case` dentro do laço, ele é recriado a cada volta: o prazo vale para cada mensagem e é renovado sempre que uma chega.
- **Timeout para a conversa inteira.** O mesmo `time.After`, mas criado uma única vez, fora do laço: o prazo vale para a conversa toda, não importa quantas mensagens cheguem.
- **Canal de parada (quit channel).** O gerador faz cada envio disputar com um canal `quit`. Quando quem consome não quer mais valores, fecha o `quit` e o gerador termina em vez de ficar bloqueado para sempre. Repare que o gerador fecha a saída ao sair, e é drenando a saída até o fechamento que a função `canalDeParada` tem certeza de que ele terminou.

```go
package main

import (
	"fmt"
	"time"
)

// tagarela é um gerador que fala cada vez mais devagar: a pausa entre as
// mensagens cresce 100ms a cada envio. Ele só para quando o canal quit é
// fechado; ao sair, fecha o canal de saída.
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

// canalDeParada lê três mensagens e manda o gerador parar fechando o canal
// quit. Em seguida drena a saída até ela ser fechada, o que garante que a
// goroutine do gerador terminou de fato.
func canalDeParada() {
	quit := make(chan struct{})
	c := tagarela("Caio", quit)

	for range 3 {
		fmt.Println(<-c)
	}
	close(quit)
	for range c {
	}
	fmt.Println("Caio parou.")
}

func main() {
	timeoutPorMensagem()
	timeoutDaConversa()
	canalDeParada()
}
```

## 🛑 Cancelamento e vazamento de goroutines

Uma _goroutine_ bloqueada em um canal que ninguém mais vai ler (ou escrever) nunca termina: ela vaza. _Goroutines_ não são coletadas pelo coletor de lixo; a memória e os recursos que elas seguram ficam presos até o fim do programa. Em um programa curto isso passa despercebido, em um servidor que roda por meses é um vazamento de memória.

O gerador `sequenciaNumeros`, usado em vários exemplos, tem esse problema: ele só termina se alguém ler todos os valores. No [exemplo](./cancelamento/cancelamento.go), a função principal lê apenas os três primeiros e para; a _goroutine_ fica presa no envio do quarto valor, como mostra a contagem de `runtime.NumGoroutine()`.

A solução é a mesma do canal de parada visto em [select](#️-select-timeout-e-quit-channel): cada envio disputa, em um `select`, com um sinal de cancelamento. Em vez de um canal `quit` próprio, o idioma em Go é receber um `context.Context` e observar `ctx.Done()`, um canal que é fechado quando o contexto é cancelado. A vantagem é que o mesmo contexto atravessa várias funções e etapas de um _pipeline_, carrega prazos (`context.WithTimeout`) e cancela todo mundo de uma vez. Para se aprofundar, veja o repositório sobre [context](https://github.com/cassiobotaro/contexto) e a segunda metade do artigo sobre [_pipelines_](https://go.dev/blog/pipelines).

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
	// Sem cancelamento: lemos só os 3 primeiros valores e paramos.
	valores := sequenciaNumeros(1, 1000)
	for range 3 {
		fmt.Printf("valor: %v\n", <-valores)
	}
	// Ninguém mais vai ler de `valores`: a goroutine do gerador está presa
	// em `saida <- 4` e continuará assim até o programa terminar.
	fmt.Printf("goroutines presas: %d\n", runtime.NumGoroutine()-1)

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
}
```

## 🚧 Trabalhador (worker)

**Também conhecido como:** consumidor, _sink_. O segundo só vale quando o trabalhador é o último estágio, isto é, quando não repassa nada adiante.

Um trabalhador é uma _goroutine_ que recebe valores de um canal e os processa.

No exemplo, valores inteiros são enviados pela função principal (main) através do canal de entrada e processados por um trabalhador.

É possível criar vários trabalhadores para processarem um mesmo canal.

Repare que o término é sinalizado com `close(pronto)`, e não com o envio de um valor: fechar um canal é o idioma em Go para comunicar um evento que acontece uma única vez, e funciona para qualquer número de leitores.

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

## 📨 Requisição e resposta

**Também conhecido como:** canal de resposta, _RPC_ interno, _restoring sequencing_ (o nome que Rob Pike dá a um uso específico da ideia, comentado abaixo).

Canais são valores como qualquer outro, então uma mensagem pode carregar um canal. Quem envia uma requisição inclui nela o canal pelo qual quer receber a resposta e fica bloqueado lendo desse canal. Quem atende processa e responde no canal que veio na mensagem. Nenhum estado é compartilhado: pedido e resposta viajam por canais. Este é o mecanismo por trás de uma _goroutine_ que funciona como serviço, e reaparece na [goroutine dona do estado](#-goroutine-dona-do-estado).

No exemplo, a função principal envia cinco requisições ao `servico` e espera cada resposta antes de enviar a próxima. O campo `resposta` é declarado como `chan<- int`: o serviço só pode escrever nele.

Na palestra [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide), Pike usa a mesma ideia para "restaurar a sequência" de um fan-in: cada mensagem carrega um canal `wait`, e quem produziu só envia a próxima mensagem depois que o leitor sinaliza nesse canal que terminou de processar a anterior.

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

## 👷‍♂️👷‍♀️ Grupo de Trabalhadores (pool of workers)

**Também conhecido como:** _worker pool_, _pool_ de _goroutines_.

A piscina de marmotinhas (carinhosamente chamada pela minha esposa) é uma coleção de _goroutines_ que ficam esperando tarefas serem atribuídas a elas. Quando a _goroutine_ finaliza a tarefa que foi atribuída, se torna disponível novamente para execução de uma nova tarefa.

No exemplo, um grupo de n trabalhadores aguarda a chegada de valores pelo canal de entrada. Cada trabalhador executa seu processamento e envia o resultado por um canal.

O grupo de trabalhadores é uma aplicação de [fan-out](#-fan-out): várias _goroutines_ leem do mesmo canal de entrada e cada valor é processado por exatamente uma delas. O que o grupo acrescenta à distribuição é o ciclo de vida dos trabalhadores, que voltam a ficar disponíveis ao terminar uma tarefa, e a coleta dos resultados em um canal de saída.

Um `sync.WaitGroup` é utilizado para saber quando todos os trabalhadores terminaram: cada trabalhador chama `wg.Done()` ao sair e uma _goroutine_ aguarda em `wg.Wait()` para então fechar o canal de saída.

Por que um `WaitGroup` e não um canal? Canais orquestram o fluxo de dados entre _goroutines_, e é isso que `entrada` e `saida` fazem aqui. Contar quantas _goroutines_ já terminaram é um problema menor, e para problemas menores Rob Pike recomenda o pacote `sync`: na palestra [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide) ele avisa "_Don't overdo it_" (às vezes só é preciso um contador) e "_Always use the right tool for the job_"; nos [Go Proverbs](https://go-proverbs.github.io/) a mesma ideia aparece como "_Channels orchestrate; mutexes serialize_".

```go
package main

import (
	"fmt"
	"sync"
)

// trabalhador processa valores recebidos do canal de entrada e envia resultados para o canal de saída.
// Ele avisa o WaitGroup quando terminar.
func trabalhador(id int, entrada <-chan int, saida chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for valor := range entrada {
		fmt.Printf("id: %d processou valor: %v\n", id, valor)
		saida <- valor * 2
	}

	fmt.Printf("id: %d terminou\n", id)
}

func grupoDeTrabalhadores(entrada <-chan int, nTrabalhadores int) chan int {
	saida := make(chan int)
	// Os canais transportam os dados; o WaitGroup apenas conta
	// quantos trabalhadores ainda não terminaram.
	var wg sync.WaitGroup

	// Cria e inicia os trabalhadores
	wg.Add(nTrabalhadores)
	for i := range nTrabalhadores {
		go trabalhador(i+1, entrada, saida, &wg)
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

> **É possível fazer só com canais.** Um canal com buffer de tamanho `n` e um laço que lê `n` vezes fazem o mesmo papel: cada trabalhador envia um sinal ao terminar e a _goroutine_ que fecha a saída espera receber todos. Funciona, mas é um `WaitGroup` reimplementado à mão. No trecho abaixo, `trabalhador` é a mesma função sem o parâmetro `wg`.
>
> ```go
> terminar := make(chan struct{}, nTrabalhadores)
>
> for i := range nTrabalhadores {
> 	go func() {
> 		trabalhador(i+1, entrada, saida)
> 		terminar <- struct{}{}
> 	}()
> }
>
> go func() {
> 	for range nTrabalhadores {
> 		<-terminar
> 	}
> 	close(saida)
> }()
> ```

## 🧑‍🏭 Pipeline

**Também conhecido como:** cadeia de estágios; cada função do _pipeline_ é um _estágio_ (_stage_).

Um _pipeline_ trabalha recebendo valores de um canal e escrevendo em outro canal, normalmente após realizar alguma transformação no valor.

No exemplo temos a função `dobro` atuando como um _pipeline_, que irá receber os valores enviados ao canal de entrada retornando os valores transformados.

Um canal pode ser definido como sendo apenas para leitura (`<-chan`) ou apenas para escrita (`chan<-`).

Os valores gerados pelo gerador `sequenciaNumeros` são enviados para o canal de entrada do pipeline e seu valor transformado recebido pelo canal de saída na função principal e é impresso.

Vários pipelines poderiam ser encadeados para realizar múltiplas transformações.

> **Atenção:** o gerador e as etapas deste _pipeline_ não são canceláveis: se o consumidor parar de ler antes do fim, a _goroutine_ vaza. Veja [Cancelamento](#-cancelamento-e-vazamento-de-goroutines).

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

**Também conhecido como:** _merge_, multiplexação (o termo que Rob Pike usa na palestra de 2012).

Um fan-in copia dados de múltiplos canais de entrada e escreve em um único canal de saída. Normalmente um fan-in só termina quando todos os canais de entrada são fechados.

A função fan-in pode receber vários canais de entrada através de [parâmetros múltiplos](https://gobyexample.com/variadic-functions).

No exemplo abaixo, enviamos vários geradores como entrada para a função fan-in e nos é retornado um único canal de saída. Internamente, uma _goroutine_ é criada para ler os valores de cada canal de entrada, porém todas escrevem no mesmo canal de saída.

Envio de mensagem em um canal fechado causa um erro (_panic_), por isso é importante garantir que todos os canais de entrada estejam fechados antes de fechar o canal de saída. Utilizamos um `sync.WaitGroup` para saber quando todos os canais de entrada foram processados, pelo mesmo motivo explicado no [grupo de trabalhadores](#️️-grupo-de-trabalhadores-pool-of-workers): os canais transportam os dados, o `WaitGroup` apenas conta quem terminou.

Repare que temos uma _goroutine_ que aguarda em `wg.Wait()` até que todas as entradas sejam consumidas, finalizando assim o canal de saída.

> **Atenção:** estes geradores não são canceláveis: se o consumidor parar de ler antes do fim, a _goroutine_ vaza. Veja [Cancelamento](#-cancelamento-e-vazamento-de-goroutines).

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

	wg.Add(len(entradas))
	for _, c := range entradas {
		go func(c <-chan int) {
			// Notifica que este canal foi processado
			defer wg.Done()
			for valor := range c {
				saida <- valor
			}
		}(c)
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

### Fan-in com uma _goroutine_ e `select`

Quando o número de entradas é fixo e conhecido, Rob Pike mostra na palestra [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide) uma variante mais enxuta: uma única _goroutine_ com um `select`, que repassa para a saída o valor da entrada que estiver pronta primeiro.

A versão da palestra roda para sempre. [Aqui](./fan_in/fan_in_select.go) ela também trata o fechamento das entradas, com o mesmo truque do **canal nil** usado na [janela deslizante](#-janela-deslizante): quando uma entrada é fechada, a variável vira `nil` e aquele `case` deixa de ser escolhido. Quando todas viram `nil`, o laço termina e a saída é fechada. Como só uma _goroutine_ escreve na saída, ela mesma fecha o canal, sem `WaitGroup`.

Quando usar cada uma? Se o número de canais é variável (um _slice_, parâmetros múltiplos), use uma _goroutine_ por entrada: um `select` tem um número fixo de `case`s escrito no código. Se o número é fixo e pequeno, o `select` é mais direto: uma _goroutine_ só e nenhuma contagem de término.

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

## 📣 Fan-out

**Também conhecido como:** distribuição, _work distribution_.

Um fan-out distribui os valores de um canal de entrada entre várias _goroutines_. O artigo sobre [_pipelines_](https://go.dev/blog/pipelines) define assim: múltiplas funções lendo do mesmo canal até que ele seja fechado. Cada valor é processado por exatamente uma delas, o que permite dividir um trabalho demorado entre vários trabalhadores.

Não é preciso nenhum código para decidir quem recebe o quê: o próprio canal faz a distribuição. Quando várias _goroutines_ estão bloqueadas lendo o mesmo canal, cada envio é entregue a apenas uma, a que estiver livre.

No exemplo, três trabalhadores dividem entre si os dez valores gerados por `sequenciaNumeros`. Repare na saída que nenhum valor aparece duas vezes. Um `sync.WaitGroup` aguarda o término de todos, pelo motivo explicado no [grupo de trabalhadores](#️️-grupo-de-trabalhadores-pool-of-workers), que é uma aplicação deste padrão.

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
func trabalhador(id int, entrada <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
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

	wg.Add(n)
	for i := range n {
		go trabalhador(i+1, entrada, &wg)
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

## 🔀 Tee (broadcast)

**Também conhecido como:** _broadcast_, _publish/subscribe_ em memória. O segundo é aproximado: em um _pub/sub_ os assinantes costumam entrar e sair dinamicamente, enquanto o tee tem um conjunto fixo de saídas.

Um tee copia cada valor de um canal de entrada para todos os canais de saída: todos os consumidores veem todos os valores. O nome vem do comando `tee` do Unix, que duplica o que recebe. É o oposto do [fan-out](#-fan-out), em que cada valor vai para um único consumidor.

No exemplo, uma sequência de números é gerada e copiada para múltiplos canais de saída. Estes canais possuem seus respectivos trabalhadores que irão fazer o processamento do valor.

O tee lê cada valor da entrada e o envia, em sequência, para cada uma das saídas; quando a entrada é fechada, fecha todas as saídas. Para aguardar o término dos trabalhadores, a função principal usa um `sync.WaitGroup`, pelo motivo explicado no [grupo de trabalhadores](#️️-grupo-de-trabalhadores-pool-of-workers).

Como os canais não têm buffer, o tee só passa para o próximo valor depois que todas as saídas receberam o atual. A consequência é que um consumidor lento atrasa todos os outros, e também o produtor: é a [contrapressão](#-contrapressão-backpressure) aplicada ao broadcast. Ninguém perde mensagem, mas o conjunto anda no ritmo do mais lento.

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
func trabalhador(id int, entrada <-chan int, demora time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()
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
	wg.Add(2)
	go trabalhador(1, saida1, 0, &wg)
	go trabalhador(2, saida2, 0, &wg)

	// Copia a sequência de números para todos os canais de saída
	tee(sequenciaNumeros(1, 10), saida1, saida2)
	wg.Wait()

	// Tee com timeout (veja tee_timeout.go): agora o trabalhador 2 é mais lento
	// do que o timeout, então parte dos valores destinados a ele é descartada.
	saida1 = make(chan int)
	saida2 = make(chan int)

	wg.Add(2)
	go trabalhador(1, saida1, 0, &wg)
	go trabalhador(2, saida2, 250*time.Millisecond, &wg)

	teeComTimeout(sequenciaNumeros(1, 5), 100*time.Millisecond, saida1, saida2)
	wg.Wait()
}
```

### Tee com timeout

Se um consumidor lento não pode segurar os demais, uma alternativa é desistir do envio depois de um tempo. [Nesta variante](./tee/tee_timeout.go), cada envio é feito dentro de um `select` que disputa com `time.After`: o que acontecer primeiro vence. Se o tempo esgotar, o valor é descartado apenas para aquela saída e o tee segue em frente. Um `select` por saída dentro do laço é suficiente, não é preciso criar uma _goroutine_ para cada envio.

Descartar mensagens é uma decisão de projeto, não parte do padrão: o consumidor lento deixa de ver todos os valores, que era justamente a garantia do tee. Por isso o descarte é registrado na saída em vez de acontecer em silêncio. Repare também que o timeout limita o atraso, mas não o elimina: cada valor ainda pode esperar até `timeout` por saída lenta. Outras formas de lidar com um consumidor lento aparecem na [janela deslizante](#-janela-deslizante) e na [contrapressão](#-contrapressão-backpressure).

No exemplo, a função principal executa as duas versões: na segunda, o trabalhador 2 leva 250ms por valor e o timeout é de 100ms, então parte dos valores destinados a ele é descartada.

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

## 🪟 Janela deslizante

**Também conhecido como:** _drop-oldest buffer_. Evite tratar _ring buffer_ como sinônimo: o _ring buffer_ é um mecanismo de armazenamento, a janela deslizante é a política de descarte (sai o mais antigo).

Uma janela deslizante (sliding window) é utilizada para prevenir que um leitor lento trave um escritor rápido. Ela funciona deslizando sobre os dados. A ordem de entregas é garantida, porém dados antigos podem ser descartados se o consumidor for muito lento.

No exemplo, uma sequência de números é gerada, porém nosso consumidor é mais lento que o produtor, logo à medida que a janela desliza os valores antigos são descartados.

Para fazer a janela deslizante, uma única _goroutine_ é dona de todo o estado (uma fila com tamanho máximo fixo) e usa um `select` para reagir ao que acontecer primeiro: se chega um valor da entrada, ele entra na fila — descartando o mais antigo se ela estiver cheia; se o consumidor está pronto para receber, o primeiro da fila é enviado.

O truque idiomático aqui é o **canal nil**: um `select` nunca escolhe um case cujo canal é `nil`. Quando a fila está vazia, o canal de envio fica `nil` e o case de envio é desabilitado (não há o que enviar); quando a entrada é fechada, a variável `entrada` é definida como `nil` e o case de recebimento é desabilitado, restando apenas drenar a fila.

Como só uma _goroutine_ toca a fila, não existe disputa entre produtor e consumidor pelo estado — uma versão anterior deste exemplo usava um canal com buffer compartilhado por duas _goroutines_ e continha uma corrida sutil que podia travar o programa.

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
func janelaDeslizante(saida chan<- int, entrada <-chan int, tamanho int) {
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
		case val, ok := <-entrada:
			if !ok {
				// Entrada fechada: desabilita este case (canal nil)
				// e continua apenas drenando a fila.
				entrada = nil
				continue
			}
			if len(fila) == tamanho {
				// Janela cheia, descarta o mais antigo e adiciona o novo
				fmt.Printf("Janela Deslizante: Buffer cheio, descartou %v para adicionar %v.\n", fila[0], val)
				fila = fila[1:]
			}
			fila = append(fila, val)

		case envio <- cabeca:
			fmt.Printf("Janela Deslizante: Enviou %v para o consumidor.\n", cabeca)
			fila = fila[1:]
		}
	}
}

// O resto do código permanece o mesmo.
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

func leitorLento(in <-chan int, pronto chan<- struct{}) {
	for val := range in {
		fmt.Printf("Consumidor: Recebeu %v\n", val)
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
	janelaDeslizante(saida, valores, 3)
	<-pronto
	fmt.Println("Fim da execução.")
}
```

## 🚦 Contrapressão (backpressure)

**Também conhecido como:** _backpressure_, _bounded queue_ (fila limitada).

Contrapressão (backpressure) é o mecanismo pelo qual um consumidor lento faz o produtor diminuir o ritmo, em vez de deixar o trabalho se acumular sem limite. É o oposto da janela deslizante: lá o produtor segue livre e os valores antigos são descartados; aqui nada é descartado, o produtor é que espera.

Em Go esse mecanismo já vem embutido nos canais. Um envio em um canal sem buffer bloqueia até que alguém leia. Um envio em um canal com buffer bloqueia assim que o buffer enche. Ou seja, a capacidade do canal define a folga máxima entre produtor e consumidor, e o bloqueio propaga a lentidão do consumidor para trás, etapa por etapa, até chegar em quem gera os dados.

No exemplo, o produtor gera dez valores o mais rápido que consegue e o consumidor leva 200ms para processar cada um. A fila entre eles tem capacidade 3. Os primeiros valores entram de imediato, mas a partir do momento em que a fila enche, cada envio leva cerca de 200ms, que é justamente o ritmo do consumidor. O produtor não tem nenhum código para "esperar o consumidor": ele apenas escreve no canal.

Repare no que não acontece: a memória não cresce, pois a fila tem um teto conhecido, e nenhum valor é perdido. O custo é que o produtor fica bloqueado, e isso precisa ser aceitável para quem está na ponta. Se quem produz é um _handler_ HTTP, por exemplo, bloquear pode significar segurar a conexão do cliente.

> **Quando bloquear não é opção.** Se o produtor não pode esperar, a alternativa é rejeitar o trabalho quando a fila está cheia usando um `select` com `default`: o envio é tentado e, se não for possível de imediato, a chamada retorna um erro (um servidor devolveria algo como `503` ou `429`). Isso é descarte de carga (load shedding), e a diferença para a janela deslizante é quem sai perdendo: na janela é o valor mais antigo, no descarte de carga é o valor novo, que nem chega a entrar. Escolher entre bloquear, descartar o antigo ou rejeitar o novo depende do que o seu sistema pode tolerar. Para limitar a taxa ao longo do tempo, e não o tamanho da fila, veja o [sistema de ticket](#-sistema-de-ticket).

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

## 🧑‍🤝‍🧑 Processamento em lote (batch processing)

**Também conhecido como:** _batching_, _micro-batching_.

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

func processadorLotes(entrada <-chan []req) chan struct{} {
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

**Também conhecido como:** _rate limiting_, _throttling_. São aproximados: aqui a taxa é fixa, sem o saldo para rajadas de um _token bucket_ (veja a nota sobre rajada abaixo).

Um sistema de ticket é usado para controlar quando um determinado trabalho pode ser executado, normalmente é utilizado para limitar o uso de um recurso sobre um período de tempo.

Exemplo: Uma API pode ser acionada apenas 15 vezes em um período de 15 minutos. A utilização é medida em blocos de 15 minutos.

No exemplo, a bilheteria é um sistema de ticket que garante que apenas 10 "tickets" sejam emitidos a cada segundo.

Enviamos através de um canal 31 processamentos a serem feitos, mas o sistema de ticket garante que apenas 10 processamentos sejam executados por segundo.

Como pode ser visto, o trabalhador pega um trabalho e fica bloqueado até que um ticket seja enviado através do canal. A ordem importa: o trabalho é lido primeiro, assim, quando o canal de trabalhos é fechado, o trabalhador encerra sem gastar um ticket à toa.

> **Nota sobre rajada (burst).** Esta implementação emite um ticket a cada `timeout/nTickets`, garantindo o teto mesmo se o consumidor for mais lento do que o ticker. Em troca, ela **não permite rajadas**: não há um saldo inicial de `nTickets` para ser consumido de uma só vez. Se você precisar de rate-limit com tolerância a rajadas (token bucket — rajada de até N seguida de reposição a `T/N`), use [`golang.org/x/time/rate`](https://pkg.go.dev/golang.org/x/time/rate).

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

func trabalhador(tickets <-chan ticket, work <-chan Trabalho) {
	for {
		// Lê o trabalho primeiro: se o canal foi fechado, encerra
		// sem gastar um ticket.
		w, ok := <-work
		if !ok {
			return // canal de trabalhos fechado
		}
		<-tickets // espera autorização antes de executar
		w()       // executa um trabalho
	}
}

// bilheteria emite, no máximo, nTickets por intervalo `timeout` —
// um ticket a cada `timeout/nTickets`. Garante o teto mesmo com consumidor
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
