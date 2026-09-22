# ⚗️ Fan-in

**Também conhecido como:** _merge_, multiplexação (o termo que Rob Pike usa na palestra de 2012).

Um fan-in copia dados de múltiplos canais de entrada e escreve em um único canal de saída. Normalmente um fan-in só termina quando todos os canais de entrada são fechados.

A função fan-in recebe os canais de entrada como [parâmetros múltiplos](https://gobyexample.com/variadic-functions).

No exemplo, três geradores são passados para a função fan-in, que devolve um único canal de saída. Por dentro há uma _goroutine_ por canal de entrada, e todas escrevem no mesmo canal de saída.

Escrever em um canal fechado causa um _panic_, então a saída só pode ser fechada depois que todas as entradas terminarem. Um `sync.WaitGroup` conta quantas ainda faltam.

Repare na _goroutine_ que espera em `wg.Wait()` e fecha a saída quando a última entrada acaba.

O `context.Context` vai para os geradores e para o próprio fan-in. Cada _goroutine_ do fan-in envia dentro de um `select` com `ctx.Done()`, então se o consumidor cancelar elas saem em vez de ficarem presas no envio, e o `WaitGroup` chega a zero do mesmo jeito. Sem isso, cancelar os geradores não bastaria: as _goroutines_ do fan-in ficariam bloqueadas com um valor na mão que ninguém vai ler.

O exemplo inteiro está em [`fan_in.go`](./fan_in.go) e o teste em [`exemplo_test.go`](./exemplo_test.go).

## Fan-in com uma _goroutine_ e `select`

Quando o número de entradas é fixo e conhecido, Rob Pike mostra na palestra [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide) uma variante mais enxuta. Uma única _goroutine_ com um `select` repassa para a saída o valor da entrada que estiver pronta primeiro.

A versão da palestra roda para sempre. [Aqui](./fan_in_select.go) ela também trata o fechamento das entradas, com o truque do canal `nil`. Um `select` nunca escolhe um `case` cujo canal é `nil`, então atribuir `nil` à variável do canal desliga aquele `case` enquanto o laço continua rodando. Quando uma entrada é fechada, a variável vira `nil` e aquele `case` deixa de ser escolhido. Quando todas viram `nil`, o laço termina e a saída é fechada. Como só uma _goroutine_ escreve na saída, ela mesma fecha o canal, sem `WaitGroup`. O truque é uma das três técnicas da palestra [Advanced Go Concurrency Patterns](https://go.dev/talks/2013/advconc.slide), de Sameer Ajmani (2013), que tem um slide só para ele.

Quando usar cada uma? Se o número de canais é variável, como em um _slice_ ou em parâmetros múltiplos, use uma _goroutine_ por entrada, porque um `select` tem um número fixo de `case`s escrito no código. Se o número é fixo e pequeno, o `select` é mais direto. Basta uma _goroutine_, e não é preciso contar quem terminou.

A variante está em [`fan_in_select.go`](./fan_in_select.go).
