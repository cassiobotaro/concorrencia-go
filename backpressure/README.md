# 🚦 Contrapressão (backpressure)

**Também conhecido como:** _backpressure_, _bounded queue_ (fila limitada).

Contrapressão (_backpressure_) é o mecanismo pelo qual um consumidor lento faz o produtor diminuir o ritmo, em vez de deixar o trabalho se acumular sem limite. Aqui nada é descartado, e quem espera é o produtor. A resposta oposta é a da [janela deslizante](../janelas_deslizantes/README.md), no fim desta parte, em que o produtor segue livre e os valores antigos são descartados.

Em Go esse mecanismo já vem embutido nos canais. A capacidade do canal é a folga máxima entre produtor e consumidor. Quando ela acaba, o envio bloqueia. O bloqueio propaga a lentidão do consumidor para trás, etapa por etapa, até chegar em quem gera os dados.

O buffer tira a sincronização entre quem envia e quem recebe, e por isso pede mais cuidado. Os exemplos daqui usam canais sem buffer sempre que podem. O buffer só aparece quando ele é a própria ideia do padrão, como aqui, no [semáforo](../semaforo/README.md) e no [primeiro a responder](../primeiro/README.md).

No exemplo, o produtor gera dez valores o mais rápido que consegue e o consumidor leva 200ms para processar cada um. A fila entre eles tem capacidade 3. Os primeiros valores entram de imediato, mas a partir do momento em que a fila enche, cada envio leva cerca de 200ms, que é justamente o ritmo do consumidor. O produtor não tem nenhum código para "esperar o consumidor", ele apenas escreve no canal.

Para tornar a espera visível, o produtor faz antes uma tentativa com `select` e `default`, que é a forma de perguntar "dá para enviar agora?" sem bloquear. Se não der, ele avisa que a fila está cheia e faz o envio bloqueante normal. É o mesmo mecanismo da nota abaixo sobre descarte de carga, mas aqui ele só observa e não descarta nada. Sem o `select` o comportamento seria o mesmo, só que em silêncio.

Repare em duas coisas que não acontecem. A memória não cresce, pois a fila tem um teto conhecido, e nenhum valor é perdido. O custo é que o produtor fica bloqueado, e isso precisa ser aceitável para quem está na ponta. Se quem produz é um _handler_ HTTP, por exemplo, bloquear pode significar segurar a conexão do cliente.

> **Quando bloquear não é opção.** Se o produtor não pode esperar, a alternativa é rejeitar o trabalho quando a fila está cheia, com um `select` e `default`. Se o envio não for possível de imediato, a chamada retorna um erro (um servidor devolveria algo como `503` ou `429`). Isso é descarte de carga (_load shedding_). A diferença para a janela deslizante é quem sai perdendo. Na janela é o valor mais antigo. No descarte de carga é o valor novo, que nem chega a entrar. Escolher entre bloquear, descartar o antigo ou rejeitar o novo depende do que o seu sistema pode tolerar. Para limitar a taxa ao longo do tempo, e não o tamanho da fila, veja o [sistema de ticket](../ticket/README.md).

O exemplo inteiro está em [`backpressure.go`](./backpressure.go) e o teste em [`exemplo_test.go`](./exemplo_test.go).
