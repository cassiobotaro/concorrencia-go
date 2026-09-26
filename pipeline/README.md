# 🏭 Pipeline

**Também conhecido como:** cadeia de estágios. Cada função do _pipeline_ é um _estágio_ (_stage_).

Um _pipeline_ recebe valores de um canal e escreve em outro, normalmente depois de transformar o valor.

No exemplo, a função `dobro` é um estágio: lê os valores do canal de entrada e escreve os valores dobrados no canal de saída.

Repare nas assinaturas. A função `dobro` recebe um `<-chan int` e devolve outro `<-chan int`. Com os tipos direcionais, o compilador impede que um estágio leia do canal em que só deveria escrever, ou escreva naquele em que só deveria ler, e fica claro quem lê e quem escreve em cada estágio. Todos os exemplos usam tipos direcionais nas assinaturas.

`sequenciaNumeros` envia os valores para o canal de entrada do _pipeline_. A função principal recebe os valores transformados pelo canal de saída e os imprime.

Os estágios podem ser encadeados. No exemplo, `dobro` é aplicado duas vezes, e cada valor sai multiplicado por quatro.

O mesmo `context.Context` atravessa o gerador e os dois estágios. Cada um faz o envio dentro de um `select` com `ctx.Done()`, como o [gerador](../geradores/README.md). Sem isso, um consumidor que parasse de ler no meio deixaria três gorrotinas presas, uma por etapa, cada uma bloqueada no envio para a seguinte. Com um único `cancel()` todas saem, e cada uma fecha a própria saída ao sair. É o que o artigo sobre [_pipelines_](https://go.dev/blog/pipelines) chama de cancelamento explícito.

> **Erros no pipeline.** O exemplo não trata erro, e há três formas de fazer isso. Parar no primeiro: o estágio devolve um segundo canal, um `<-chan error` com buffer 1, escreve o erro nele e retorna. O buffer é para a gorrotina não ficar presa no envio se ninguém estiver lendo o erro ainda. Carregar o erro junto do valor: a saída vira um `struct{ valor int; err error }`, e quem lê decide o que fazer com cada um. É o que o [primeiro a responder](../primeiro/README.md) faz. Juntar todos sem parar: o estágio recebe um canal de erros como parâmetro e segue processando. A escolha depende de um erro dever ou não interromper os demais estágios. Quando deve, o `errgroup.WithContext` da [variante do semáforo](../semaforo/README.md#e-com-errgroup) faz o primeiro erro cancelar o contexto que já atravessa os estágios.

O exemplo inteiro está em [`pipeline.go`](./pipeline.go) e o teste em [`exemplo_test.go`](./exemplo_test.go).
