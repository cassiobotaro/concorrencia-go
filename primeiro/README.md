# 🏁 Primeiro a responder

**Também conhecido como:** _hedged request_, réplicas, `First` (o nome da função na palestra de Pike).

Para não depender do servidor mais lento, envie a mesma requisição a várias réplicas e use a primeira resposta que chegar. É a técnica que Rob Pike usa no exemplo da busca do Google, na palestra [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide), para reduzir a latência de cauda. Combinada com um prazo em `context.WithTimeout`, o resultado é o que Pike descreve como um programa rápido, replicado e robusto.

A palestra é de 2012, e o `First` de Pike só lê a primeira resposta. As perdedoras continuam trabalhando até o fim e jogam o resultado fora. Em uma chamada de rede, isso é uma requisição a mais no servidor por réplica. No [exemplo](./primeiro.go), `primeiro` recebe um `context.Context`, deriva um filho com `context.WithCancel`, passa esse filho a cada réplica e cancela ao retornar. As perdedoras desistem no próximo `ctx.Done()`. Uma réplica de verdade faria o mesmo com `http.NewRequestWithContext`, que interrompe a requisição inteira quando o contexto é cancelado. É o cancelamento por `context`, assunto do README de [cancelamento](../cancelamento/README.md), aplicado ao padrão.

Repare no canal com buffer de tamanho `len(replicas)`. Mesmo com o cancelamento, uma perdedora pode terminar entre a chegada da vencedora e o `cancel()`. Com um canal sem buffer ela ficaria bloqueada no envio para sempre, pois ninguém mais vai ler. Com uma vaga por réplica, ela deposita a resposta e termina.

Repare também que uma réplica que falha envia o erro pelo mesmo canal, em vez de sair calada. O `First` de Pike não tem erro, então não precisa disso. Aqui, se as falhas fossem ignoradas e todas as réplicas falhassem, ninguém enviaria nada e `primeiro` esperaria para sempre, pois o `ctx.Done()` de um `context.Background()` nunca chega. Por isso o laço espera uma resposta por réplica: a primeira sem erro vence, e se todas falharem o último erro é devolvido. O `exemplo_test.go` cobre esse caminho com duas réplicas que só falham.

No exemplo, as réplicas são simuladas com uma espera aleatória de até 100ms, então a vencedora muda a cada execução. A segunda parte combina `primeiro` com um prazo de 20ms. O prazo vem em um `context.WithTimeout`, e `primeiro` o repassa às réplicas, então o mesmo `ctx.Done()` que encerra a espera encerra também as réplicas.

O exemplo inteiro está em [`primeiro.go`](./primeiro.go) e o teste em [`exemplo_test.go`](./exemplo_test.go).
