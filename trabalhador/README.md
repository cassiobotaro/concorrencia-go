# 🚧 Trabalhador (worker)

**Também conhecido como:** consumidor, _sink_. O segundo nome só vale quando o trabalhador é o último estágio, ou seja, quando não repassa nada adiante.

Um trabalhador é uma gorrotina que recebe valores de um canal e os processa até o canal ser fechado. É o papel oposto ao do [gerador](../geradores/README.md): lá uma gorrotina escreve no canal, aqui uma gorrotina lê dele.

Em Go, o trabalho cabe em um `for valor := range entrada`. O `range` recebe um valor por vez e termina sozinho quando quem envia fecha o canal. O trabalhador não precisa de contador nem de valor especial para saber que acabou.

No [exemplo](./trabalhador.go), a função principal envia os inteiros de 0 a 9 pela `entrada`, fecha o canal e espera em `<-pronto`. O trabalhador imprime cada valor. O `fmt.Printf` serve apenas para mostrar a execução e não faz parte do padrão. No lugar dele entraria o processamento de verdade.

Por que esperar em `<-pronto`? Fechar a `entrada` não espera nada. Quando o último envio retorna, o trabalhador recebeu o 9 mas talvez ainda não o tenha impresso. Quando a função principal retorna, o programa termina, e as gorrotinas que ainda estavam rodando terminam junto. Sem a espera, o exemplo perdeu o "valor: 9" em 178 de 200 execuções. Um `time.Sleep` no fim de `main` esconderia o problema sem resolvê-lo, porque nenhum tempo fixo garante que o trabalho acabou.

Repare que o término é sinalizado com `close(pronto)`, e não com o envio de um valor. Fechar um canal é a forma usual em Go de comunicar um evento que acontece uma única vez, como "terminei" ou "pode parar". Todos os que estiverem lendo o canal são desbloqueados ao mesmo tempo, então o sinal funciona para qualquer número de leitores. Por isso o canal é um `chan struct{}`, que não carrega dado nenhum. Quem fecha `pronto` é a gorrotina anônima de `main`, e não o `trabalhador`. Ele só processa valores e nem sabe que o sinal existe, o mesmo desenho do [grupo de trabalhadores](../grupo/README.md).

O custo é que o trabalhador dita o ritmo. Com o canal sem buffer, cada envio da função principal só retorna quando o trabalhador está pronto para receber, então um trabalhador lento segura quem produz. A [contrapressão](../backpressure/README.md) explica por que isso pode ser o comportamento desejado. Quando não é, vários trabalhadores podem ler do mesmo canal, e esse é o [fan-out](../fan_out/README.md), mais adiante.

O exemplo inteiro está em [`trabalhador.go`](./trabalhador.go) e o teste em [`exemplo_test.go`](./exemplo_test.go).
