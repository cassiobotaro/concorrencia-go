# 🆕 Geradores

**Também conhecido como:** produtor, _source_. É o mesmo papel do `produtor` da [contrapressão](../backpressure/README.md).

Um gerador é uma função que dispara uma gorrotina para escrever uma sequência de valores em um canal, e devolve esse canal a quem a chamou. O produtor roda em paralelo com o consumidor. Isso importa quando produzir custa caro, como ler de disco ou de rede, e quando os valores vão atravessar um [_pipeline_](../pipeline/README.md) de etapas concorrentes.

No [exemplo](./geradores.go), `sequenciaNumeros` gera mil inteiros. A função principal lê o canal e imprime os valores. Com o `range`, a iteração continua até o canal ser fechado.

Repare no `context.Context` e no `select` dentro da gorrotina. Cada envio disputa com `ctx.Done()`. Sem isso, um consumidor que parasse de ler no meio deixaria a gorrotina presa para sempre no próximo envio, com o canal e tudo o que ela segura. Com o contexto, quem consome cancela, e a gorrotina sai do laço e fecha o canal. No `main` isso não chega a acontecer, porque o laço lê os mil valores. O `exemplo_test.go` da pasta faz o outro caminho: lê um valor, cancela e drena o canal até ele fechar, o que prova que a gorrotina terminou. É a forma de escrever um gerador em código de hoje, e o README de [cancelamento](../cancelamento/README.md) diz o que acontece quando ela falta.

O custo é que a responsabilidade passa para quem consome. Ele precisa criar o contexto e cancelar ao sair, mesmo quando leu tudo, e o `defer cancelar()` do exemplo está ali por isso. Esquecer o `cancel` é um erro que o `go vet` aponta (`lostcancel`): o contexto filho fica registrado no pai até o pai ser cancelado.

A função `sequenciaNumeros` reaparece em vários exemplos, sempre com o contexto. Ela é copiada de propósito, para que cada arquivo possa ser lido e executado sozinho. Como diz um dos [Go Proverbs](https://go-proverbs.github.io/), "_a little copying is better than a little dependency_".

O exemplo inteiro está em [`geradores.go`](./geradores.go) e o teste em [`exemplo_test.go`](./exemplo_test.go).
