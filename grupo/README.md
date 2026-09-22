# 👷 Grupo de Trabalhadores (pool of workers)

**Também conhecido como:** _worker pool_, _pool_ de gorrotinas.

A piscina de marmotinhas (carinhosamente chamada pela minha esposa) é uma coleção de gorrotinas que ficam esperando tarefas serem atribuídas a elas. Quando termina uma tarefa, a gorrotina volta a ficar disponível para a próxima.

No exemplo, dois trabalhadores esperam valores no canal de entrada. Cada um dobra o valor que recebe e envia o resultado pelo canal de saída.

O grupo de trabalhadores é uma aplicação de [fan-out](../fan_out/README.md). Várias gorrotinas leem do mesmo canal e cada valor vai para uma só. Além de distribuir o trabalho, o grupo junta os resultados em um canal de saída.

O grupo fixa quantas gorrotinas existem. Se a ideia for ter uma gorrotina por tarefa e limitar apenas quantas executam ao mesmo tempo, veja o [semáforo](../semaforo/README.md). Se as tarefas devolvem erro, nos dois casos, veja a variante [com errgroup](../semaforo/README.md#e-com-errgroup).

Como no [fan-out](../fan_out/README.md), a ordem da saída muda a cada execução.

Os trabalhadores são iniciados com `wg.Go`, e outra gorrotina espera em `wg.Wait()` para fechar o canal de saída, como no [fan-in](../fan_in/README.md). Repare que o `trabalhador` nem sabe que o `WaitGroup` existe. Ele só processa valores, e quem o dispara é que cuida de esperar.

O mesmo `context.Context` vai para o gerador e para os trabalhadores. Cada trabalhador envia o resultado dentro de um `select` com `ctx.Done()`, como no [pipeline](../pipeline/README.md). Sem isso, um consumidor que parasse de ler a saída no meio deixaria os dois trabalhadores presos no envio, cada um com um resultado na mão que ninguém vai ler, e a gorrotina do `wg.Wait()` nunca fecharia a saída. Com o cancelamento, os trabalhadores saem, o `WaitGroup` chega a zero e a saída é fechada do mesmo jeito.

O exemplo inteiro está em [`grupo.go`](./grupo.go) e o teste em [`exemplo_test.go`](./exemplo_test.go).
