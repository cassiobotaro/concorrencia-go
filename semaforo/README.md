# 🚥 Semáforo (paralelismo limitado)

**Também conhecido como:** _bounded parallelism_, limite de gorrotinas em voo.

Um canal com buffer de capacidade `n` funciona como um semáforo. Enviar ocupa uma vaga, e bloqueia quando todas estão ocupadas. Receber libera uma vaga. Com isso dá para limitar quantas gorrotinas executam um trecho ao mesmo tempo sem criar um grupo fixo. Cada tarefa tem sua própria gorrotina, mas só `n` avançam de cada vez. A técnica aparece com o nome de _bounded parallelism_ no artigo sobre [_pipelines_](https://go.dev/blog/pipelines).

Qual a diferença para os padrões vizinhos? O [grupo de trabalhadores](../grupo/README.md) fixa o número de gorrotinas. O [sistema de ticket](../ticket/README.md) limita a taxa ao longo do tempo. O semáforo limita quantas tarefas executam ao mesmo tempo. Este é um dos poucos casos em que o buffer do canal não é um ajuste fino, porque a capacidade do canal é o próprio limite.

No exemplo, dez tarefas são disparadas de uma vez, mas o semáforo tem três vagas. A saída mostra que o número de tarefas ativas nunca passa de três. O contador atômico (`sync/atomic`) serve apenas para observar isso e não faz parte do padrão. O `sync.WaitGroup` aguarda o término de todas as tarefas.

O exemplo inteiro está em [`semaforo.go`](./semaforo.go) e o teste em [`exemplo_test.go`](./exemplo_test.go).

## E com errgroup?

O canal de vagas e o `WaitGroup` fazem duas coisas que o padrão sempre precisa: limitar e esperar. O pacote [`golang.org/x/sync/errgroup`](https://pkg.go.dev/golang.org/x/sync/errgroup) junta as duas em um tipo só e acrescenta a terceira, que os outros exemplos ignoram: o erro. `g.SetLimit(3)` é o canal de três vagas. `g.Go` dispara a tarefa, e bloqueia quando as vagas acabam. `g.Wait` espera todas, como o `WaitGroup`, e devolve o primeiro erro que alguma tarefa retornou. Com `errgroup.WithContext`, esse primeiro erro cancela um contexto, e as tarefas que ainda não começaram podem desistir olhando `ctx.Err()`.

A [versão abaixo](./com_errgroup.go) faz isso. Com três vagas e a tarefa 2 falhando, as tarefas 1 a 3 começam, e as outras sete são canceladas antes de começar, porque quando elas conseguem uma vaga o contexto já foi cancelado. A tarefa que falha é andaime, existe só para mostrar o cancelamento, e ela falha assim que começa, antes de fazer o trabalho. Isso é de propósito. Se ela falhasse ao terminar, as tarefas 1 e 3 terminariam no mesmo instante e poderiam liberar vaga antes de o cancelamento chegar, e a tarefa 4 começaria em parte das execuções. O `errgroup` garante que o erro cancela o contexto antes de a vaga ser devolvida, então quem entra depois de uma falha sempre a enxerga.

O custo é uma dependência fora da biblioteca padrão. É a única que o código importa. O `staticcheck` que aparece no `go.mod` é ferramenta de análise, declarado com a diretiva `tool`, e não entra no binário. Vale a pena quando as tarefas devolvem erro e uma falha deve interromper as demais, que é o caso comum em código de produção. Quando as tarefas não falham, ou quando cada erro deve ser tratado por conta própria, o canal com buffer e o `WaitGroup` bastam.

> **Vagas com peso.** O mesmo módulo do `errgroup` tem o pacote [`golang.org/x/sync/semaphore`](https://pkg.go.dev/golang.org/x/sync/semaphore). `NewWeighted(n)` cria o semáforo e `Acquire(ctx, peso)` ocupa `peso` vagas de uma vez, então uma tarefa pesada pode valer por duas ou três. A espera pela vaga respeita o contexto, o que o `sem <- struct{}{}` do exemplo só faria dentro de um `select` com `ctx.Done()`.

A variante está em [`com_errgroup.go`](./com_errgroup.go).
