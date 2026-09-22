# 🗺️ Olá Mundo

Um canal serve de ponte entre a função principal e uma gorrotina.

O programa principal fica bloqueado em `<-canal` até que a gorrotina envie a mensagem "Olá, mundo!". Quando isso acontece, ele recebe a mensagem e a exibe. Como o canal não tem buffer, o envio e o recebimento acontecem juntos: nenhum dos dois lados segue em frente sem o outro.

Quando o programa principal termina, a gorrotina termina junto. Aqui isso não faz diferença, porque ela já enviou o que tinha a enviar. Nos padrões seguintes faz, e é por isso que quase todos eles recebem um `context.Context` ou fecham um canal para avisar que acabaram.

O exemplo inteiro está em [`ola_mundo.go`](./ola_mundo.go) e o teste em [`exemplo_test.go`](./exemplo_test.go).
