# 📨 Requisição e resposta

**Também conhecido como:** canal de resposta, _RPC_ interno, _restoring sequencing_ (o nome que Rob Pike dá a um uso específico da ideia, comentado abaixo).

Canais são valores como qualquer outro, então uma mensagem pode carregar um canal. Quem envia uma requisição inclui nela o canal pelo qual quer receber a resposta e fica bloqueado lendo desse canal. Quem atende processa e responde no canal que veio na mensagem. Nenhum estado é compartilhado, pois pedido e resposta viajam por canais. É assim que se faz uma gorrotina funcionar como um serviço, e a ideia reaparece na [gorrotina dona do estado](../dono_do_estado/README.md).

No exemplo, a função principal envia cinco requisições ao `servico` e espera cada resposta antes de enviar a próxima. O campo `resposta` é declarado como `chan<- int`, então o serviço só pode escrever nele.

Repare na vaga de buffer do canal de resposta. Sem ela, o serviço ficaria preso no envio se quem pediu desistisse de esperar, por exemplo em um `select` com `ctx.Done()`, e nenhuma outra requisição seria atendida. Com a vaga, o serviço deposita a resposta e segue, e a resposta que ninguém leu vai embora com o canal. No exemplo isso não acontece, porque a função principal sempre lê, mas quem copia o padrão para um serviço de verdade precisa da vaga. O [primeiro a responder](../primeiro/README.md) usa o mesmo buffer, com uma vaga por réplica.

Na palestra [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide), Pike usa a mesma ideia para "restaurar a sequência" de um fan-in. Cada mensagem carrega um canal `wait`, e quem produziu só envia a próxima depois que o leitor avisa, por esse canal, que terminou de processar a anterior.

O exemplo inteiro está em [`requisicao_resposta.go`](./requisicao_resposta.go) e o teste em [`exemplo_test.go`](./exemplo_test.go).
