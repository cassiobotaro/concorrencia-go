# 👷 Grupo de Trabalhadores (pool of workers)

**Também conhecido como:** _worker pool_, _pool_ de _goroutines_.

A piscina de marmotinhas (carinhosamente chamada pela minha esposa) é uma coleção de _goroutines_ que ficam esperando tarefas serem atribuídas a elas. Quando termina uma tarefa, a _goroutine_ volta a ficar disponível para a próxima.

No exemplo, dois trabalhadores esperam valores no canal de entrada. Cada um dobra o valor que recebe e envia o resultado pelo canal de saída.

O grupo de trabalhadores é uma aplicação de [fan-out](../fan_out/README.md). Várias _goroutines_ leem do mesmo canal e cada valor vai para uma só. Além de distribuir o trabalho, o grupo junta os resultados em um canal de saída.

O grupo fixa quantas _goroutines_ existem. Se a ideia for ter uma _goroutine_ por tarefa e limitar apenas quantas executam ao mesmo tempo, veja o [semáforo](../semaforo/README.md). Se as tarefas devolvem erro, nos dois casos, veja a variante [com errgroup](../semaforo/README.md#e-com-errgroup).

Como no [fan-out](../fan_out/README.md), a ordem da saída muda a cada execução.

Os trabalhadores são iniciados com `wg.Go`, e outra _goroutine_ espera em `wg.Wait()` para fechar o canal de saída, como no [fan-in](../fan_in/README.md). Repare que o `trabalhador` nem sabe que o `WaitGroup` existe. Ele só processa valores, e quem o dispara é que cuida de esperar.

O exemplo inteiro está em [`grupo.go`](./grupo.go) e o teste em [`exemplo_test.go`](./exemplo_test.go).
