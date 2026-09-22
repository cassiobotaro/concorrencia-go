# 🎫 Sistema de ticket

**Também conhecido como:** _rate limiting_, _throttling_. São nomes aproximados, porque aqui a taxa é fixa, sem o saldo para rajadas de um _token bucket_ (veja a nota sobre rajada abaixo).

Um sistema de ticket controla quando um trabalho pode ser executado. Serve para limitar o uso de um recurso ao longo de um período, como uma API que aceita 15 chamadas a cada 15 minutos.

No exemplo, a bilheteria emite no máximo 10 tickets por segundo. A função principal envia 31 trabalhos pelo canal, e eles saem a 10 por segundo, levando cerca de três segundos no total.

O ticket limita a taxa ao longo do tempo. Para limitar quantas tarefas executam ao mesmo tempo, veja o [semáforo](../semaforo/README.md).

O trabalhador pega um trabalho e fica bloqueado até receber um ticket. A ordem importa. Como o trabalho é lido primeiro, o trabalhador encerra sem gastar um ticket à toa quando o canal de trabalhos é fechado.

> **Nota sobre rajada (_burst_).** Esta implementação emite um ticket a cada `janela/nTickets`, e o intervalo é contado a partir da entrega do ticket anterior, com um `time.Timer` reiniciado a cada entrega. Por isso o teto vale mesmo se o consumidor for mais lento do que a bilheteria. Um `time.Ticker` não serviria: ele guarda um tick enquanto o envio espera pelo consumidor, e esse tick faria o ticket seguinte sair na hora, dois de uma vez. Em troca, ela não permite rajadas, pois não há um saldo inicial de `nTickets` para ser consumido de uma só vez. Se você precisar de _rate limiting_ com tolerância a rajadas (_token bucket_, isto é, rajada de até N seguida de reposição a `T/N`), use [`golang.org/x/time/rate`](https://pkg.go.dev/golang.org/x/time/rate).

O exemplo inteiro está em [`ticket.go`](./ticket.go) e o teste em [`exemplo_test.go`](./exemplo_test.go).
