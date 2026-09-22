# Concorrência em Go

Go é fundamentada no modelo CSP (Communicating sequential processes) proposto por Tony Hoare. Neste modelo, os dados são compartilhados enviando mensagens por canais.

No CSP original de Hoare, um processo envia mensagens direto para outro, identificado pelo nome. Erlang seguiu esse caminho. Go vem de outro ramo da família, o das linguagens Newsqueak, Alef e Limbo. Nelas o canal é um valor de primeira classe, que pode ser guardado em variáveis, passado como parâmetro e até enviado por outro canal. Os dois modelos são equivalentes, só se expressam de forma diferente. Rob Pike compara com arquivos: em Erlang é como escrever em um arquivo pelo nome, em Go é como escrever por meio de um descritor de arquivo.

Outra ideia que volta várias vezes é que [concorrência não é paralelismo](https://go.dev/blog/waza-talk). Concorrência é compor computações independentes, uma forma de estruturar o programa. Paralelismo é executá-las ao mesmo tempo. Um programa concorrente pode rodar em um único processador. E um programa bem estruturado para concorrência costuma paralelizar bem quando há mais de um.

O texto parte do princípio de que você conhece Go, inclusive _goroutines_, canais, `select` e `sync.WaitGroup`. Se alguma dessas peças for nova, passe antes pelo [Tour of Go](https://go.dev/tour/concurrency/1), pelos capítulos de concorrência do [Go by Example](https://gobyexample.com/goroutines), que cobrem canais com e sem buffer, direção, `select`, timeouts, fechamento de canal e `WaitGroup`, e pela seção de concorrência do [Effective Go](https://go.dev/doc/effective_go#concurrency). Os padrões daqui usam essas peças sem explicá-las de novo.

Cada parte vai do mais simples ao mais complexo. A [Parte 1](#parte-1--padrões-básicos) traz os padrões básicos, que são a forma dos canais entre as _goroutines_. A [Parte 2](#parte-2--controlando-o-ritmo) trata de produtores e consumidores em ritmos diferentes. A [Parte 3](#parte-3--conversa-entre-goroutines) é sobre uma _goroutine_ responder a outra. A [Parte 4](#parte-4--encerrar-e-supervisionar) ensina a encerrar _goroutines_ e a vigiar as que continuam. Para ir mais fundo em `context`, sugiro também [este repositório](https://github.com/cassiobotaro/contexto).

Os `fmt.Print` dos exemplos estão ali só para você enxergar a execução e não fazem parte dos padrões. Em código real seriam _logs_, ou nem existiriam.

Cada pasta é um programa independente, que você executa com `go run ./pipeline/`, e tem um `README.md` com a explicação do padrão. Cada uma tem também um `exemplo_test.go` com uma função `Example` que confere a saída. Ela usa `// Output:` quando a ordem das linhas é fixa e `// Unordered output:` quando só o conjunto é previsível. Para rodar tudo com o detector de corrida, use `go test -race ./...`.

Os mesmos padrões aparecem com outros nomes em livros, artigos e outras linguagens, por isso cada seção traz uma linha "Também conhecido como". Dois pedem cuidado: "produtor" e "consumidor" são papéis, não padrões, e quase todo exemplo tem os dois. Aparecem como nomes alternativos de [Geradores](./geradores/README.md) e [Trabalhador](./trabalhador/README.md) porque nessas seções cada papel aparece sozinho.

## Parte 1 · Padrões básicos

Os blocos de montar. Cada padrão desta parte faz uma coisa só. Os padrões das partes seguintes são combinações e variações destes.

- [🆕 Geradores](./geradores/README.md): uma _goroutine_ produz uma sequência de valores em um canal e para quando o contexto é cancelado.
- [🚧 Trabalhador (worker)](./trabalhador/README.md): uma _goroutine_ consome valores de um canal e avisa, fechando outro, quando terminou.
- [🏭 Pipeline](./pipeline/README.md): estágios encadeados por canais, cada um transformando o que recebe.
- [📣 Fan-out](./fan_out/README.md): vários trabalhadores leem do mesmo canal e cada valor vai para um só.
- [🔀 Tee (broadcast)](./tee/README.md): cada valor é copiado para todas as saídas, e uma variante desiste do envio por timeout.
- [⚗️ Fan-in](./fan_in/README.md): vários canais de entrada viram um só, com uma _goroutine_ por entrada ou um `select`.
- [👷 Grupo de Trabalhadores (pool of workers)](./grupo/README.md): um número fixo de _goroutines_ divide a entrada e junta os resultados em uma saída.

## Parte 2 · Controlando o ritmo

O que fazer quando quem produz e quem consome andam em velocidades diferentes. Cada padrão desta parte dá uma resposta: fazer o produtor esperar, limitar quantos executam ao mesmo tempo, limitar a taxa, agrupar o trabalho ou descartar o que ficou velho.

- [🚦 Contrapressão (backpressure)](./backpressure/README.md): a capacidade do canal faz o produtor esperar o consumidor lento, sem descartar nada.
- [🚥 Semáforo (paralelismo limitado)](./semaforo/README.md): um canal com buffer limita quantas _goroutines_ executam ao mesmo tempo, e o `errgroup` faz o mesmo com erros.
- [🎫 Sistema de ticket](./ticket/README.md): um ticker regula quantos trabalhos executam por intervalo de tempo.
- [📦 Processamento em lote (batch processing)](./lote/README.md): itens que chegam um a um são agrupados por tamanho, por tempo ou sob demanda.
- [🪟 Janela deslizante](./janelas_deslizantes/README.md): uma fila de tamanho fixo descarta o valor mais antigo quando o consumidor não acompanha.

## Parte 3 · Conversa entre goroutines

Nos padrões anteriores, o canal só carrega dados em um sentido. Nesta parte a mensagem leva junto um canal de resposta, ou esperar a resposta faz parte do padrão. É assim que uma _goroutine_ vira um serviço, que um estado ganha uma dona e que a mesma pergunta pode ser feita a várias réplicas de uma vez.

- [📨 Requisição e resposta](./requisicao_resposta/README.md): a mensagem carrega o canal em que quem pediu espera a resposta.
- [🔐 Goroutine dona do estado](./dono_do_estado/README.md): uma única _goroutine_ guarda o estado e atende pedidos por canais, com a versão em mutex ao lado.
- [🏁 Primeiro a responder](./primeiro/README.md): a mesma consulta vai a várias réplicas, vale a primeira resposta e as outras são canceladas.

## Parte 4 · Encerrar e supervisionar

Os geradores das partes anteriores já recebem um `context.Context` e saem quando ele é cancelado. O motivo é o vazamento. Uma _goroutine_ bloqueada em um canal que ninguém mais vai ler nunca termina, o coletor de lixo não recolhe _goroutines_, e a memória e os recursos que ela segura ficam presos até o fim do programa. Em um servidor que roda por meses, isso é um vazamento de memória.

Desde o Go 1.26 dá para encontrar essas _goroutines_. O perfil `goroutineleak`, do pacote `runtime/pprof`, lista as que estão bloqueadas em um canal ou mutex que nenhuma _goroutine_ viva alcança. É experimental, atrás de `GOEXPERIMENT=goroutineleakprofile`, e aparece também em `/debug/pprof/goroutineleak`. Para ir mais fundo em `context`, veja [este repositório](https://github.com/cassiobotaro/contexto) e a segunda metade do artigo sobre [_pipelines_](https://go.dev/blog/pipelines).

As duas palestras que mais aparecem aqui, a de Rob Pike (2012) e a de Sameer Ajmani (2013), são anteriores ao pacote `context`, que só entrou na biblioteca padrão no Go 1.7, em 2016. Foi o próprio Ajmani quem o apresentou, no [post](https://go.dev/blog/context) de julho de 2014. O que o `context` padronizou foi uma única técnica das palestras: o canal `quit`, fechado para avisar todo mundo de uma vez. É o `ctx.Done()`. O resto continua sem substituto, porque o contexto leva o sinal em um sentido só, de quem chama para quem é chamado, e nunca traz resultado de volta. O laço `for` com `select` e estado local, o canal de resposta que confirma a parada com um erro e o canal `nil` que desliga um `case` são escritos à mão hoje do mesmo jeito que em 2013.

Esta parte trata do que vem depois de mandar parar, e do que fazer quando ninguém mandou: como saber que a _goroutine_ parou, como juntar vários motivos de parada em um só e como descobrir que um trabalhador travou sem avisar. Mostra a forma com `context`, que é a que você vai encontrar em código de hoje, e cita a das palestras onde ela ajuda a entender o que o `context` faz por dentro.

- [🤝 Parada com confirmação](./cancelamento/README.md#-parada-com-confirmação): quem manda parar espera a _goroutine_ liberar recursos, com `cancel()` e `g.Wait()`.
- [🧩 Combinar sinais de parada (or-channel)](./cancelamento/README.md#-combinar-sinais-de-parada-or-channel): vários motivos de parada viram um canal só, quando não dá para derivar contextos.
- [💓 Heartbeat](./batimento/README.md): um trabalhador emite sinal de vida e o supervisor decide que ele morreu quando o sinal para.

## Curiosidade

Um exemplo que não resolve nenhum problema do dia a dia, mas mostra o quanto uma _goroutine_ é barata.

- [⛓️ Daisy-chain](./corrente/README.md): 10 mil _goroutines_ em corrente, para mostrar quanto uma _goroutine_ custa.

## Referências

- [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide), Rob Pike, Google I/O 2012.
- [Advanced Go Concurrency Patterns](https://go.dev/talks/2013/advconc.slide), Sameer Ajmani, Google I/O 2013.
- [Go Concurrency Patterns: Context](https://go.dev/blog/context), Sameer Ajmani, 2014.
- [Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines), Sameer Ajmani, 2014.
- [Concurrency is not parallelism](https://go.dev/blog/waza-talk), Rob Pike, 2012.
- [Go Proverbs](https://go-proverbs.github.io/), Rob Pike, Gopherfest 2015.
- [Concurrency in Go](https://www.oreilly.com/library/view/concurrency-in-go/9781491941294/), Katherine Cox-Buday, O'Reilly, 2017.
- [Apresentação sobre concorrência](https://github.com/andrebq/andrebq.github.io) do @andrebq, de onde vêm boa parte das explicações e dos exemplos.
- [Tour of Go](https://go.dev/tour/concurrency/1), [Go by Example](https://gobyexample.com/goroutines) e [Effective Go](https://go.dev/doc/effective_go#concurrency), para as peças da linguagem.
- [golang.org/x/sync/errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup) e [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate).
- [Repositório sobre context](https://github.com/cassiobotaro/contexto), do mesmo autor.
