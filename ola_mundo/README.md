# 🗺️ Olá mundo

**Também conhecido como:** o _hello world_ dos canais, o primeiro exemplo do [Tour of Go](https://go.dev/tour/concurrency/2) e do [Go by Example](https://gobyexample.com/channels).

Um canal liga a função principal a uma gorrotina. É o menor programa concorrente que faz alguma coisa: uma gorrotina envia, a função principal recebe.

No exemplo, a função principal fica bloqueada em `<-canal` até a gorrotina enviar "Olá, mundo!". Como o canal não tem buffer, o envio e o recebimento acontecem juntos: nenhum dos dois lados segue em frente sem o outro. Repare que não há `WaitGroup` nem `time.Sleep` para esperar a gorrotina. O recebimento já é a espera.

Quando a função principal retorna, o programa acaba e a gorrotina morre junto. Aqui isso não faz diferença, porque ela já enviou o que tinha a enviar. Nos padrões seguintes faz, e é por isso que quase todos recebem um `context.Context`, como os [geradores](../geradores/README.md), ou fecham um canal para avisar que acabaram, como o [trabalhador](../trabalhador/README.md).

O exemplo inteiro está em [`ola_mundo.go`](./ola_mundo.go) e o teste em [`exemplo_test.go`](./exemplo_test.go).
