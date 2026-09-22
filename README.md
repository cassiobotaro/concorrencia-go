# Concorrência em Go

Go é fundamentada no modelo CSP (Communicating sequential processes) proposto por Tony Hoare. Neste modelo, os dados são compartilhados enviando mensagens por canais.

No CSP original de Hoare, um processo envia mensagens direto para outro, identificado pelo nome. Erlang seguiu esse caminho. Go vem de outro ramo da família, o das linguagens Newsqueak, Alef e Limbo. Nelas o canal é um valor de primeira classe, que pode ser guardado em variáveis, passado como parâmetro e até enviado por outro canal. Os dois modelos são equivalentes, só se expressam de forma diferente. Rob Pike compara com arquivos: em Erlang é como escrever em um arquivo pelo nome, em Go é como escrever por meio de um descritor de arquivo.

Outra ideia que volta várias vezes é que [concorrência não é paralelismo](https://go.dev/blog/waza-talk). Concorrência é compor computações independentes, uma forma de estruturar o programa. Paralelismo é executá-las ao mesmo tempo. Um programa concorrente pode rodar em um único processador. E um programa bem estruturado para concorrência costuma paralelizar bem quando há mais de um.

O texto parte do princípio de que você conhece Go, inclusive gorrotinas, canais, `select` e `sync.WaitGroup`. Se alguma dessas peças for nova, passe antes pelo [Tour of Go](https://go.dev/tour/concurrency/1), pelos capítulos de concorrência do [Go by Example](https://gobyexample.com/goroutines), que cobrem canais com e sem buffer, direção, `select`, timeouts, fechamento de canal e `WaitGroup`, e pela seção de concorrência do [Effective Go](https://go.dev/doc/effective_go#concurrency). Os padrões daqui usam essas peças sem explicá-las de novo.

Para ir mais fundo em `context`, sugiro também [este repositório](https://github.com/cassiobotaro/contexto).

Os `fmt.Print` dos exemplos estão ali só para você enxergar a execução e não fazem parte dos padrões. Em código real seriam _logs_, ou nem existiriam.

Cada pasta é um programa independente, que você executa com `go run ./pipeline/`, e tem um `README.md` com a explicação do padrão. Cada uma tem também um `exemplo_test.go` que confere a saída. Quando o exemplo não depende do relógio, o teste é uma função `Example`, que usa `// Output:` quando a ordem das linhas é fixa e `// Unordered output:` quando só o conjunto é previsível. Quando o exemplo dorme, espera um _ticker_ ou tem prazo, o teste roda dentro de `synctest.Test`, do pacote `testing/synctest`. Lá dentro, na chamada bolha, o relógio é falso e só anda quando todas as gorrotinas estão bloqueadas. Um `time.Sleep` de um segundo termina na hora, e o tempo medido é sempre o mesmo. `synctest.Test` pede um `*testing.T`, que uma `Example` não recebe, então esses testes são funções `Test` e capturam a saída com o pacote [`internal/saida`](./internal/saida/saida.go). A bolha ainda confere uma coisa a mais: se sobrar gorrotina viva quando o teste acaba, ele falha. Para rodar tudo com o detector de corrida, use `go test -race ./...`.

Os mesmos padrões aparecem com outros nomes em livros, artigos e outras linguagens, por isso o README de cada padrão traz uma linha "Também conhecido como". Dois pedem cuidado: "produtor" e "consumidor" são papéis, não padrões, e quase todo exemplo tem os dois. Aparecem como nomes alternativos de [Geradores](./geradores/README.md) e [Trabalhador](./trabalhador/README.md) porque nesses dois cada papel aparece sozinho.

## Padrões

Do mais simples ao mais complexo. Os primeiros são a forma dos canais entre as gorrotinas; depois vêm o controle de ritmo, a conversa entre gorrotinas e, por fim, como encerrar e supervisionar.

- [🗺️ Olá Mundo](./ola_mundo/README.md): uma gorrotina envia uma mensagem por um canal e a função principal a recebe.
- [🆕 Geradores](./geradores/README.md): uma gorrotina produz uma sequência de valores em um canal e para quando o contexto é cancelado.
- [🚧 Trabalhador (worker)](./trabalhador/README.md): uma gorrotina consome valores de um canal e avisa, fechando outro, quando terminou.
- [🏭 Pipeline](./pipeline/README.md): estágios encadeados por canais, cada um transformando o que recebe.
- [📣 Fan-out](./fan_out/README.md): vários trabalhadores leem do mesmo canal e cada valor vai para um só.
- [🔀 Tee (broadcast)](./tee/README.md): cada valor é copiado para todas as saídas, e uma variante desiste do envio por timeout.
- [⚗️ Fan-in](./fan_in/README.md): vários canais de entrada viram um só, com uma gorrotina por entrada ou um `select`.
- [👷 Grupo de Trabalhadores (pool of workers)](./grupo/README.md): um número fixo de gorrotinas divide a entrada e junta os resultados em uma saída.
- [🚦 Contrapressão (backpressure)](./backpressure/README.md): a capacidade do canal faz o produtor esperar o consumidor lento, sem descartar nada.
- [🚥 Semáforo (paralelismo limitado)](./semaforo/README.md): um canal com buffer limita quantas gorrotinas executam ao mesmo tempo, e o `errgroup` faz o mesmo com erros.
- [🎫 Sistema de ticket](./ticket/README.md): um ticker regula quantos trabalhos executam por intervalo de tempo.
- [📦 Processamento em lote (batch processing)](./lote/README.md): itens que chegam um a um são agrupados por tamanho, por tempo ou sob demanda.
- [🪟 Janela deslizante](./janelas_deslizantes/README.md): uma fila de tamanho fixo descarta o valor mais antigo quando o consumidor não acompanha.
- [📨 Requisição e resposta](./requisicao_resposta/README.md): a mensagem carrega o canal em que quem pediu espera a resposta.
- [🔐 Gorrotina dona do estado](./dono_do_estado/README.md): uma única gorrotina guarda o estado e atende pedidos por canais, com a versão em mutex ao lado.
- [🏁 Primeiro a responder](./primeiro/README.md): a mesma consulta vai a várias réplicas, vale a primeira resposta e as outras são canceladas.
- [🤝 Parada com confirmação](./cancelamento/README.md#-parada-com-confirmação): quem manda parar espera a gorrotina liberar recursos, com `cancel()` e `g.Wait()`.
- [🧩 Combinar sinais de parada (or-channel)](./cancelamento/README.md#-combinar-sinais-de-parada-or-channel): vários motivos de parada viram um canal só, quando não dá para derivar contextos.
- [💓 Heartbeat](./batimento/README.md): um trabalhador emite sinal de vida e o supervisor decide que ele morreu quando o sinal para.
- [⛓️ Daisy-chain](./corrente/README.md): uma curiosidade, 10 mil gorrotinas em corrente para mostrar quanto uma gorrotina custa.

## Referências

- [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide), Rob Pike, Google I/O 2012.
- [Advanced Go Concurrency Patterns](https://go.dev/talks/2013/advconc.slide), Sameer Ajmani, Google I/O 2013.
- [Go Concurrency Patterns: Context](https://go.dev/blog/context), Sameer Ajmani, 2014.
- [Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines), Sameer Ajmani, 2014.
- [Concurrency is not parallelism](https://go.dev/blog/waza-talk), Rob Pike, 2012.
- [Testing concurrent code with testing/synctest](https://go.dev/blog/synctest), Damien Neil, 2025.
- [Testing Time (and other asynchronicities)](https://go.dev/blog/testing-time), Damien Neil, 2025.
- [Go Proverbs](https://go-proverbs.github.io/), Rob Pike, Gopherfest 2015.
- [Concurrency in Go](https://www.oreilly.com/library/view/concurrency-in-go/9781491941294/), Katherine Cox-Buday, O'Reilly, 2017.
- [Apresentação sobre concorrência](https://github.com/andrebq/andrebq.github.io) do @andrebq, de onde vêm boa parte das explicações e dos exemplos.
- [Tour of Go](https://go.dev/tour/concurrency/1), [Go by Example](https://gobyexample.com/goroutines) e [Effective Go](https://go.dev/doc/effective_go#concurrency), para as peças da linguagem.
- [golang.org/x/sync/errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup) e [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate).
- [Repositório sobre context](https://github.com/cassiobotaro/contexto), do mesmo autor.
