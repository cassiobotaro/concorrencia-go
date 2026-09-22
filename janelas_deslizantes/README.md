# 🪟 Janela deslizante

**Também conhecido como:** _drop-oldest buffer_. Evite tratar _ring buffer_ como sinônimo. O _ring buffer_ é uma forma de armazenar os dados, enquanto a janela deslizante é a regra de descarte, em que sai sempre o mais antigo.

Uma janela deslizante (_sliding window_) impede que um leitor lento trave um escritor rápido. É a resposta oposta à da [contrapressão](../backpressure/README.md): em vez de o produtor esperar, os valores mais velhos são descartados. A ordem das entregas é mantida, mas um consumidor lento perde os valores que já saíram da janela.

No exemplo, o produtor envia um valor por segundo e o consumidor leva quatro segundos para processar cada um. A janela guarda três valores, então, à medida que ela desliza, os mais antigos são descartados.

Para fazer a janela deslizante, uma única gorrotina é dona de todo o estado (uma fila com tamanho máximo fixo) e usa um `select` para reagir ao que acontecer primeiro. Se chega um valor da entrada, ele entra na fila, e o mais antigo é descartado caso ela esteja cheia. Se o consumidor está pronto para receber, o primeiro da fila é enviado.

O truque aqui é o canal `nil`, visto no [fan-in com select](../fan_in/README.md#fan-in-com-uma-gorrotina-e-select). Como um `case` cujo canal é `nil` nunca é escolhido, dá para ligar e desligar cada `case` conforme o estado da fila. Quando a fila está vazia, o canal de envio fica `nil` e o `case` de envio é desabilitado, pois não há o que enviar. Quando a entrada é fechada, a variável `entrada` passa a valer `nil` e o `case` de recebimento é desabilitado. Daí em diante só resta esvaziar a fila.

Como só uma gorrotina toca a fila, não há disputa entre produtor e consumidor pelo estado. É a técnica da [gorrotina dona do estado](../dono_do_estado/README.md). Uma versão anterior deste exemplo usava um canal com buffer compartilhado por duas gorrotinas e tinha uma corrida sutil que podia travar o programa.

O exemplo inteiro está em [`janelas_deslizantes.go`](./janelas_deslizantes.go) e o teste em [`exemplo_test.go`](./exemplo_test.go).
