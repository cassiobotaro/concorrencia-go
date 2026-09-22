# 🔀 Tee (broadcast)

**Também conhecido como:** _broadcast_, _publish/subscribe_ em memória. O segundo é aproximado: em um _pub/sub_ os assinantes costumam entrar e sair dinamicamente, enquanto o tee tem um conjunto fixo de saídas.

Um tee copia cada valor de um canal de entrada para todos os canais de saída, de modo que todos os consumidores veem todos os valores. O nome vem do comando `tee` do Unix, que duplica o que recebe. É o oposto do [fan-out](../fan_out/README.md), em que cada valor vai para um único consumidor.

No exemplo, uma sequência de dez números é copiada para dois canais de saída. Cada canal tem seu trabalhador, e os dois recebem todos os valores.

O tee lê cada valor da entrada e o envia, em sequência, para cada uma das saídas. Quando a entrada é fechada, ele fecha todas as saídas. Para aguardar o término dos trabalhadores, a função principal usa um `sync.WaitGroup`.

Como os canais não têm buffer, o tee só passa para o próximo valor depois que todas as saídas receberam o atual. A consequência é que um consumidor lento atrasa todos os outros, e também o produtor. É a [contrapressão](../backpressure/README.md) aplicada ao broadcast. Ninguém perde mensagem, mas todos andam no ritmo do mais lento.

O exemplo inteiro está em [`tee.go`](./tee.go) e o teste em [`exemplo_test.go`](./exemplo_test.go).

## Tee com timeout

Se um consumidor lento não pode segurar os demais, uma alternativa é desistir do envio depois de um tempo. [Nesta variante](./tee_timeout.go), cada envio é feito dentro de um `select` que disputa com `time.After`, e vence o que acontecer primeiro. Se o tempo esgotar, o valor é descartado apenas para aquela saída e o tee segue em frente. Um `select` por saída dentro do laço é suficiente, não é preciso criar uma _goroutine_ para cada envio.

O `time.After` dentro do laço é recriado a cada envio, e o prazo vale para cada valor em cada saída. Até o Go 1.22, cada chamada deixava um timer vivo até disparar, mesmo depois de o `select` ter escolhido outro `case`, e a recomendação era evitar `time.After` em laço. Desde o Go 1.23, um timer que o programa não referencia mais é recolhido pelo coletor de lixo na hora, desde que o `go.mod` declare `go 1.23` ou mais novo, como o deste repositório. Em código real o prazo costuma chegar de fora, em um `context.Context` criado com `context.WithTimeout`, e o `case` passa a ser `<-ctx.Done()`. A diferença é que o mesmo prazo vale para a operação inteira e atravessa as funções chamadas. O [primeiro a responder](../primeiro/README.md) faz isso.

Descartar mensagens é uma decisão de projeto, e não parte do padrão. Com o descarte, o consumidor lento deixa de ver todos os valores, que era justamente a garantia do tee. Por isso o exemplo avisa na saída cada vez que descarta, em vez de descartar em silêncio. O timeout limita o atraso, mas não acaba com ele. Cada valor ainda pode esperar até `timeout` em cada saída lenta. Outras formas de lidar com um consumidor lento aparecem na [janela deslizante](../janelas_deslizantes/README.md) e na [contrapressão](../backpressure/README.md).

No exemplo, a função principal executa as duas versões. Na segunda, o trabalhador 2 leva 250ms por valor e o timeout é de 100ms, então parte dos valores destinados a ele é descartada.

A variante está em [`tee_timeout.go`](./tee_timeout.go).
