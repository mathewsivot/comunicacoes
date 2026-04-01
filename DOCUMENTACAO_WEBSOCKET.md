# Documentacao da Aplicacao

## Visao Geral

Esta aplicacao foi organizada em dois programas:

- `Server/`: servidor que recebe conexoes WebSocket, valida o handshake inicial e responde a requisicoes do cliente.
- `Client/`: cliente de terminal que abre a conexao, negocia os parametros da sessao e envia a consulta desejada.

A parte principal da aplicacao esta na comunicacao entre cliente e servidor usando `gorilla/websocket`. O foco do sistema e estabelecer uma sessao validada antes de qualquer troca de dados de negocio.

## Arquitetura Geral

O servidor sobe um endpoint HTTP em `:3000` e expoe o caminho WebSocket `ws://localhost:3000/ws`.

O cliente:

1. pede ao usuario o modo de operacao;
2. pede ao usuario o `max_message_size` desejado;
3. abre uma conexao WebSocket com o servidor;
4. envia um handshake em JSON;
5. aguarda a confirmacao do servidor;
6. envia a requisicao da moeda;
7. recebe a resposta final e exibe no terminal.

O servidor:

1. aceita o upgrade HTTP para WebSocket;
2. limita o tamanho inicial de leitura;
3. exige que a primeira mensagem seja um handshake valido;
4. negocia os parametros da sessao;
5. confirma o handshake ao cliente;
6. registra no terminal que a conexao foi estabelecida;
7. recebe a requisicao da moeda;
8. devolve a resposta em JSON.

## Papel do WebSocket

O uso de WebSocket substitui a comunicacao TCP textual anterior por um canal com mensagens estruturadas. Isso traz tres ganhos principais:

- mensagens JSON com contrato claro;
- handshake explicito antes da operacao principal;
- validacao mais simples de tipos, ordem das mensagens e limites da sessao.

O pacote `gorilla/websocket` e usado dos dois lados:

- no servidor, para fazer o `Upgrade` da requisicao HTTP e ler/escrever mensagens JSON;
- no cliente, para abrir a conexao com `Dial` e trocar mensagens JSON com o servidor.

## Fluxo de Conexao

### 1. Inicializacao do servidor

O servidor registra o handler em `/ws` e fica aguardando conexoes. Quando uma requisicao chega nesse endpoint, ele tenta converter a conexao HTTP para WebSocket.

Se o upgrade falhar, a conexao e rejeitada.

### 2. Escolha do modo no cliente

Antes de se conectar, o cliente pede que o usuario escolha um modo de operacao:

- `gbn`: Go-Back-N
- `sr`: Selective Repeat

No estado atual do projeto, esse modo faz parte da negociacao da sessao. Ou seja, ele e validado e confirmado no handshake, mesmo que ainda nao altere a logica de retransmissao internamente.

Em seguida, o cliente tambem pede o `max_message_size` da sessao. Esse valor passa a ser definido pelo proprio usuario no terminal antes da abertura da conexao.
O valor minimo aceito e `30`, enquanto o limite maximo continua sendo controlado pelo teto interno do servidor.

### 3. Handshake inicial

Assim que a conexao WebSocket e aberta, o cliente envia a primeira mensagem da sessao:

```json
{
  "type": "handshake_request",
  "operation_mode": "gbn",
  "max_message_size": 512
}
```

Essa mensagem define dois pontos obrigatorios da conexao:

- `operation_mode`: modo escolhido pelo cliente, aceitando `gbn` ou `sr`;
- `max_message_size`: tamanho maximo de mensagem que o cliente deseja usar, agora escolhido pelo usuario no terminal.

Regra atual do tamanho:

- minimo permitido no cliente e no servidor: `30`
- maximo efetivo da sessao: menor valor entre o pedido do cliente e o teto interno do servidor

### 4. Validacao do handshake no servidor

O servidor trata essa primeira mensagem como obrigatoria. Ele rejeita a conexao quando:

- a primeira mensagem nao e `handshake_request`;
- o modo de operacao nao e suportado;
- o tamanho maximo informado e invalido.

Quando o handshake e valido, o servidor:

- limita o tamanho efetivo da sessao ao menor valor entre o pedido do cliente e o limite interno do servidor;
- responde com um `handshake_response`;
- registra no terminal que a conexao foi confirmada.

Exemplo de resposta:

```json
{
  "type": "handshake_response",
  "status": "accepted",
  "operation_mode": "gbn",
  "max_message_size": 512
}
```

Exemplo de log no terminal do servidor:

```text
handshake confirmado: conexao estabelecida com 127.0.0.1:xxxxx | modo=gbn | max_message_size=512
```

Esse log deixa explicito que a conexao ocorreu e que o handshake foi concluido com sucesso.

## Protocolo de Mensagens

As mensagens foram centralizadas em um contrato compartilhado para evitar divergencia entre cliente e servidor.

### Handshake

- `handshake_request`
- `handshake_response`

### Operacao principal

- `rate_request`
- `rate_response`

### Erro de protocolo

- `error`

Todas as mensagens sao trafegadas em JSON.

## Fluxo apos o Handshake

Depois da confirmacao da sessao, o cliente pede ao usuario o codigo da moeda e envia:

```json
{
  "type": "rate_request",
  "currency": "BTC"
}
```

O servidor le essa mensagem somente depois do handshake. Se o cliente tentar enviar outra coisa antes disso, a conexao e tratada como invalida.

Quando a requisicao e aceita, o servidor devolve:

```json
{
  "type": "rate_response",
  "currency": "BTC",
  "price": 123.45
}
```

Se houver falha de processamento, o servidor responde com um campo `error`.

## Validacoes Importantes

O protocolo atual garante:

- ordem obrigatoria de mensagens;
- negociacao inicial de modo e tamanho maximo;
- rejeicao de modos nao suportados;
- rejeicao de mensagens fora do contrato;
- confirmacao visual no servidor quando o handshake conclui.

Essas validacoes impedem que o cliente entre na fase principal da sessao sem antes estabelecer um acordo minimo de comunicacao.

## Como Executar

### Servidor

```bash
cd Server
go run .
```

### Cliente

```bash
cd Client
go run .
```

Fluxo esperado no cliente:

1. escolher `gbn` ou `sr`;
2. informar o `max_message_size`;
3. aguardar a confirmacao do handshake;
4. informar a moeda;
5. receber a resposta.

Fluxo esperado no servidor:

1. iniciar o endpoint `/ws`;
2. receber a conexao;
3. validar o handshake;
4. imprimir a confirmacao da conexao no terminal;
5. processar a requisicao seguinte.

## Resumo Final

O comportamento central da aplicacao esta na abertura da sessao WebSocket e na negociacao inicial entre cliente e servidor. O cliente nao envia a operacao principal sem antes negociar o modo e o limite da conexao. O servidor, por sua vez, so aceita continuar a comunicacao quando o handshake respeita o protocolo definido.

Isso torna a comunicacao mais previsivel, mais clara para depuracao e mais facil de evoluir futuramente.
