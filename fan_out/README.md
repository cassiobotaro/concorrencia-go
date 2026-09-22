# 📣 Fan-out

**Também conhecido como:** distribuição, _work distribution_.

Um fan-out distribui os valores de um canal de entrada entre várias gorrotinas. O artigo sobre [_pipelines_](https://go.dev/blog/pipelines) define assim: múltiplas funções lendo do mesmo canal até que ele seja fechado. Cada valor é processado por exatamente uma delas, o que permite dividir um trabalho demorado entre vários trabalhadores.

Não é preciso nenhum código para decidir quem recebe o quê, porque o próprio canal faz a distribuição. Quando várias gorrotinas estão bloqueadas lendo o mesmo canal, cada envio é entregue a apenas uma delas.

No exemplo, três trabalhadores dividem entre si os dez valores gerados por `sequenciaNumeros`. Repare na saída que nenhum valor aparece duas vezes. Um `sync.WaitGroup` aguarda o término de todos. O [grupo de trabalhadores](../grupo/README.md), mais adiante, é uma aplicação deste padrão.

Por que um `WaitGroup` e não um canal? Canais servem para orquestrar o fluxo de dados entre gorrotinas. Contar quantas já terminaram é um problema menor, e para esses Rob Pike recomenda o pacote `sync`. Na palestra [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide) ele avisa "_Don't overdo it_", porque às vezes só é preciso um contador. É o provérbio "_Channels orchestrate; mutexes serialize_", dos [Go Proverbs](https://go-proverbs.github.io/). Por isso, em todos os exemplos daqui, os canais transportam os dados e o `WaitGroup` apenas conta quem terminou. O `wg.Go` existe desde o Go 1.25 e faz o `Add` e o `Done` de uma vez. Em código mais antigo você vai encontrar `wg.Add(1)` antes de cada `go` e `defer wg.Done()` dentro da gorrotina.

Execute o exemplo mais de uma vez e veja que a ordem da saída muda. Os trabalhadores concorrem pelos valores da entrada, e quem decide qual deles roda a cada momento é o escalonador. É a primeira vez que o não determinismo aparece por aqui, e ele vem da ideia de que [concorrência não é paralelismo](https://go.dev/blog/waza-talk): o programa descreve computações independentes, mas não diz em que ordem elas executam. Por isso um programa concorrente correto não pode depender dessa ordem.

Não confunda com o [tee](../tee/README.md), em que cada valor é copiado para todos os consumidores.

O exemplo inteiro está em [`fan_out.go`](./fan_out.go) e o teste em [`exemplo_test.go`](./exemplo_test.go).
