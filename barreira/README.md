# 🛂 Barreira

**Também conhecido como:** _barrier_, _rendezvous_ (quando são só duas gorrotinas).

Várias gorrotinas fazem uma parte do trabalho, param em um ponto e nenhuma segue até a última chegar. É o que se quer quando a segunda fase depende de todas as primeiras: carregar os dados em paralelo e só então começar a calcular, por exemplo. Com duas gorrotinas o nome é _rendezvous_, o encontro marcado.

Em Go a barreira é um `sync.WaitGroup` com `Add(n)`. Cada gorrotina chama `Done` ao chegar e `Wait` em seguida. Nos outros exemplos deste material só a função principal chama `Wait`, e as gorrotinas só chamam `Done`. Aqui cada uma faz os dois: é esperada pelas outras e espera as outras. `Wait` pode ser chamado por quantas gorrotinas for preciso, e todas acordam juntas quando o contador chega a zero.

No [exemplo](./barreira.go), quatro gorrotinas chegam à barreira a cada 50ms, e a última chega aos 150ms. Só então as quatro passam. A saída mostra as quatro chegadas em ordem e, depois de todas elas, as quatro passagens, em ordem imprevisível. O `time.Sleep` é andaime, existe para escalonar as chegadas e deixar visível que ninguém passa antes.

O [semáforo](../semaforo/README.md) limita quantas gorrotinas passam por vez. A barreira faz o contrário: exige que todas cheguem para alguma passar. O `wg.Wait()` da função principal em [fan-out](../fan_out/README.md) é uma barreira com um único lado esperando. Fechar um canal também libera todo mundo de uma vez, é o que `ctx.Done()` faz no [cancelamento](../cancelamento/README.md), mas alguém precisa decidir a hora de fechar. Na barreira a última a chegar libera as outras sem saber que é a última.

> **Uma vez só.** O `WaitGroup` não volta a `n` sozinho, então esta barreira serve para uma rodada. Para repetir a cada rodada, é um `WaitGroup` novo por rodada ou um `sync.Cond`, cujo `Broadcast` acorda todas as gorrotinas paradas em `Wait`. O `sync.Cond` não aparece neste material, e o livro [Go Concurrency Distilled](https://antonz.org/go-concurrency-distilled/), de onde vem este padrão, tem um capítulo sobre ele.

O exemplo inteiro está em [`barreira.go`](./barreira.go) e o teste em [`exemplo_test.go`](./exemplo_test.go).
