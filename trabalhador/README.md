# 🚧 Trabalhador (worker)

**Também conhecido como:** consumidor, _sink_. O segundo só vale quando o trabalhador é o último estágio, isto é, quando não repassa nada adiante.

Um trabalhador é uma _goroutine_ que recebe valores de um canal e os processa.

No exemplo, a função principal envia dez valores inteiros pelo canal de entrada, e um trabalhador os processa.

Vários trabalhadores podem ler do mesmo canal. É o [fan-out](../fan_out/README.md), mais adiante.

Repare que o término é sinalizado com `close(pronto)`, e não com o envio de um valor. Fechar um canal é a forma usual em Go de comunicar um evento que acontece uma única vez, como "terminei" ou "pode parar". Funciona para qualquer número de leitores, porque todos os que estiverem lendo são desbloqueados ao mesmo tempo. Por isso o canal é um `chan struct{}`, que não carrega dado nenhum.

O exemplo inteiro está em [`trabalhador.go`](./trabalhador.go) e o teste em [`exemplo_test.go`](./exemplo_test.go).
