# 🆕 Geradores (generators)

**Também conhecido como:** produtor, _source_. É o mesmo papel do `produtor` da [contrapressão](../backpressure/README.md).

Um gerador é uma função que dispara uma gorrotina para escrever uma sequência de valores em um canal, e devolve esse canal a quem a chamou. O produtor roda ao mesmo tempo que o consumidor. Isso importa quando produzir custa caro, como ler de disco ou de rede, e quando os valores vão atravessar um [_pipeline_](../pipeline/README.md) de etapas concorrentes.

A saída tentadora é devolver uma fatia. Ela não serve: o consumidor só vê o primeiro valor depois que o último foi produzido, e a fatia inteira fica na memória. Com o canal sem buffer, o gerador anda no máximo um valor à frente de quem lê.

No [exemplo](./geradores.go), `sequenciaNumeros` gera mil inteiros. A função principal lê o canal com `range` e imprime os valores. O `range` termina quando a gorrotina fecha o canal, no `defer close(saida)`.

Repare no `context.Context` e no `select` dentro da gorrotina. Cada envio disputa com `ctx.Done()`. Sem isso, um consumidor que parasse de ler no meio deixaria a gorrotina presa para sempre no próximo envio, com o canal e tudo o que ela segura. Com o contexto, quem consome cancela, e a gorrotina sai do laço e fecha o canal. O README de [cancelamento](../cancelamento/README.md) mostra o que acontece quando o contexto falta.

No `main` o cancelamento não chega a acontecer, porque o laço lê os mil valores. O [`exemplo_test.go`](./exemplo_test.go) faz o outro caminho: lê um valor, cancela e drena o canal até ele fechar. Se a gorrotina não terminasse, o canal nunca fecharia e o teste ficaria preso.

O custo é que a responsabilidade passa para quem consome. Ele precisa criar o contexto e cancelar ao sair, mesmo quando leu tudo, e o `defer cancelar()` do exemplo está ali por isso. Esquecer de chamar `cancelar` é um erro que o `go vet` aponta (`lostcancel`). Quando o pai pode ser cancelado, o contexto filho fica registrado nele até um dos dois ser cancelado.

A função `sequenciaNumeros` reaparece em outros seis exemplos, sempre com o contexto. Ela é copiada de propósito, para que cada arquivo possa ser lido e executado sozinho. Como diz um dos [Go Proverbs](https://go-proverbs.github.io/), "_a little copying is better than a little dependency_".

O exemplo inteiro está em [`geradores.go`](./geradores.go) e o teste em [`exemplo_test.go`](./exemplo_test.go).
