# 🛑 Cancelamento

Os geradores dos outros exemplos já recebem um `context.Context` e saem quando ele é cancelado. O motivo é o vazamento. Uma _goroutine_ bloqueada em um canal que ninguém mais vai ler nunca termina, o coletor de lixo não recolhe _goroutines_, e a memória e os recursos que ela segura ficam presos até o fim do programa. Em um servidor que roda por meses, isso é um vazamento de memória.

Desde o Go 1.26 dá para encontrar essas _goroutines_. O perfil `goroutineleak`, do pacote `runtime/pprof`, lista as que estão bloqueadas em um canal ou mutex que nenhuma _goroutine_ viva alcança. É experimental, atrás de `GOEXPERIMENT=goroutineleakprofile`, e aparece também em `/debug/pprof/goroutineleak`. Para ir mais fundo em `context`, veja [este repositório](https://github.com/cassiobotaro/contexto) e a segunda metade do artigo sobre [_pipelines_](https://go.dev/blog/pipelines).

As duas palestras que mais aparecem neste material, a de Rob Pike (2012) e a de Sameer Ajmani (2013), são anteriores ao pacote `context`, que só entrou na biblioteca padrão no Go 1.7, em 2016. Foi o próprio Ajmani quem o apresentou, no [post](https://go.dev/blog/context) de julho de 2014. O que o `context` padronizou foi uma única técnica das palestras: o canal `quit`, fechado para avisar todo mundo de uma vez. É o `ctx.Done()`. O resto continua sem substituto, porque o contexto leva o sinal em um sentido só, de quem chama para quem é chamado, e nunca traz resultado de volta. O laço `for` com `select` e estado local, o canal de resposta que confirma a parada com um erro e o canal `nil` que desliga um `case` são escritos à mão hoje do mesmo jeito que em 2013.

Os dois padrões desta pasta tratam do que vem depois de mandar parar: como saber que a _goroutine_ parou e como juntar vários motivos de parada em um só. O [heartbeat](../batimento/README.md) cuida do caso em que ninguém mandou. Os exemplos mostram a forma com `context`, que é a que você vai encontrar em código de hoje, e citam a das palestras onde ela ajuda a entender o que o `context` faz por dentro.

Os dois padrões dividem o mesmo `main`, em [`cancelamento.go`](./cancelamento.go).

## 🤝 Parada com confirmação

**Também conhecido como:** _shutdown_ com _ack_, _graceful stop_.

Mandar "pare" não garante que a _goroutine_ já parou. Se ela precisa liberar recursos antes de sair (fechar arquivos, encerrar conexões), quem pediu a parada deve esperar a confirmação. Nas palestras isso era feito à mão. Rob Pike, em [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide), reaproveita o canal `quit`: quem quer parar envia "pare", a _goroutine_ faz a limpeza e responde no mesmo canal. Sameer Ajmani, em [Advanced Go Concurrency Patterns](https://go.dev/talks/2013/advconc.slide), usa [requisição e resposta](../requisicao_resposta/README.md): o método `Close` envia um canal de resposta por um `chan chan error` e espera nele, e a _goroutine_ faz a limpeza e responde com o erro, se houver. Nos dois, o pedido desce e a resposta sobe pelo mesmo mecanismo.

O `context` faz só a metade de baixo. Ele leva o sinal de quem chama para quem é chamado e nunca traz nada de volta. Para a metade de cima, o costume hoje é juntar o `context` com um `errgroup`, o mesmo pacote visto na [variante do semáforo](../semaforo/README.md#e-com-errgroup). A _goroutine_ roda dentro de `g.Go`, faz a limpeza ao ver `ctx.Done()` e devolve o erro. Quem quer parar chama `cancel()` e depois `g.Wait()`, que bloqueia até a _goroutine_ retornar e entrega esse erro. É o `Close` de Ajmani em duas chamadas, sem canal de resposta escrito à mão. O [exemplo](./context_errgroup.go) faz isso: o gerador libera os recursos, e só então o programa segue. O `time.Sleep` dentro de `limpeza` é andaime, está ali para a liberação levar tempo visível.

Quando a confirmação precisa carregar mais do que um erro, o canal de resposta continua sendo a forma. `context.WithCancelCause`, que existe desde o Go 1.20, deixa quem cancela dizer o motivo, lido do outro lado com `context.Cause(ctx)`. Isso é informação descendo, de quem cancela para quem é cancelado, e não substitui a confirmação.

O exemplo está em [`context_errgroup.go`](./context_errgroup.go), e [`cancelamento.go`](./cancelamento.go) tem o `main` que roda os dois exemplos da pasta.

## 🧩 Combinar sinais de parada (or-channel)

**Também conhecido como:** _or-channel_. Não confunda com o _or-done-channel_, do mesmo livro citado abaixo, que é outra técnica. Ele embrulha a leitura de um canal para que ela também respeite um sinal de parada.

Às vezes uma _goroutine_ deve parar quando _qualquer um_ de vários sinais chegar: o contexto da requisição, um sinal do sistema operacional, um prazo global. Em 2017, quando o livro citado abaixo saiu, cada um desses era um canal diferente, e a resposta era combinar os canais em um só. Hoje os três são contextos. O da requisição sempre foi, um _handler_ HTTP recebe `r.Context()`. O prazo é `context.WithTimeout`. E desde o Go 1.16, `signal.NotifyContext` devolve um contexto cancelado quando o sinal do sistema chega. Quando as origens são contextos, a resposta é derivar um do outro. O filho é cancelado quando o pai é, e a _goroutine_ fica com um único `case`, o `ctx.Done()`. A primeira metade do [exemplo](./qualquer.go) faz isso com as três origens.

O que sobra para combinar à mão é o que não é contexto: o canal `pronto` de outra _goroutine_, a saída de um gerador, qualquer `chan struct{}` fechado como sinal. A função `qualquer` combina esses canais em um só, que é fechado quando o primeiro deles fechar. Na segunda metade do exemplo, a _goroutine_ para quando o contexto acabar ou quando um colega terminar, o que vier primeiro.

A implementação usa uma _goroutine_ por canal de entrada, e a primeira a ser acordada fecha a saída. Ela toma dois cuidados. O primeiro é o `sync.Once`, que garante um único `close` mesmo que dois sinais cheguem juntos, já que fechar um canal duas vezes causa _panic_. O segundo é que cada _goroutine_ também observa a própria saída. Assim, quando um sinal vence, as demais terminam em vez de vazarem esperando canais que talvez nunca fechem.

Existem alternativas. Para duas ou três origens, um `select` explícito é o mais claro. Para o caso geral há a versão recursiva, que divide a lista ao meio, e o `reflect.Select`. As duas funcionam, mas são mais engenhosas do que claras, e os Go Proverbs lembram que "_Clear is better than clever_" e "_Reflection is never clear_". Para o sentido inverso, ligar um contexto a algo que não entende contextos, `context.AfterFunc(ctx, f)` roda `f` quando o contexto é cancelado, desde o Go 1.21.

De onde vem isso? Sinalizar a parada fechando um canal aparece no artigo sobre [_pipelines_](https://go.dev/blog/pipelines), com o canal `done`, e na palestra [Advanced Go Concurrency Patterns](https://go.dev/talks/2013/advconc.slide), de Sameer Ajmani (2013), cujo código fecha um canal `quit` para encerrar as _goroutines_ do `Merge`. Nenhum dos dois combina vários sinais em um só. O _or-channel_, com esse nome, é do livro _Concurrency in Go_, de Katherine Cox-Buday (O'Reilly, 2017, capítulo 4), que usa a versão recursiva.

O exemplo está em [`qualquer.go`](./qualquer.go), e [`cancelamento.go`](./cancelamento.go) tem o `main` que roda os dois exemplos da pasta.
