# Documentacao da Aplicacao

## Visao Geral

Esta aplicacao e composta por dois programas:

- `Server/`: servidor UDP que recebe fragmentos de mensagem, valida checksum, confirma recebimentos com `ACK` e solicita retransmissao com `NAK`.
- `Client/`: cliente de terminal que fragmenta uma mensagem, aplica uma criptografia simples, simula perdas/erros e envia os pacotes ao servidor.

Apesar do nome deste arquivo mencionar WebSocket, a implementacao atual do projeto usa **UDP**. A comunicacao acontece pela porta `10000`, com mensagens textuais separadas por `|`.

O objetivo principal e simular conceitos de comunicacao confiavel sobre UDP:

- fragmentacao de mensagens;
- janela deslizante;
- checksum;
- deteccao de erro;
- confirmacao com `ACK`;
- retransmissao com `NAK` e timeout;
- modos Go-Back-N e Selective Repeat;
- perdas e erros simulados.

## Arquitetura Geral

O servidor escuta em:

```text
0.0.0.0:10000
```

O cliente envia pacotes para:

```text
127.0.0.1:10000
```

Fluxo geral do cliente:

1. pede ao usuario o modo de operacao;
2. envia um `HELLO` para registrar a sessao no servidor;
3. recebe `HELLO_ACK` com o tamanho da janela;
4. pede a mensagem ao usuario;
5. divide a mensagem em fragmentos de 4 caracteres;
6. criptografa cada fragmento;
7. calcula o checksum;
8. envia pacotes `DATA`;
9. simula perda ou erro em alguns pacotes;
10. processa `ACK`, `NAK` e timeouts;
11. retransmite conforme o modo escolhido.

Fluxo geral do servidor:

1. escuta pacotes UDP na porta `10000`;
2. recebe `HELLO` e cria estado para o cliente;
3. recebe pacotes `DATA`;
4. descriptografa o payload;
5. valida o checksum;
6. responde com `ACK` ou `NAK`;
7. armazena os fragmentos recebidos;
8. remonta e imprime a mensagem completa.

## Protocolo de Mensagens

As mensagens trafegam como texto, com campos separados por `|`.

### HELLO

Enviado pelo cliente antes dos dados:

```text
HELLO|30|gobackn
```

Campos:

- `HELLO`: identifica o inicio da sessao;
- `30`: tamanho minimo esperado para a mensagem, mantido como valor informativo no codigo atual;
- `gobackn` ou `selecionado`: modo de operacao escolhido.

Resposta do servidor:

```text
HELLO_ACK|5
```

O valor `5` representa o tamanho da janela configurada no servidor.

### DATA

Enviado pelo cliente para cada fragmento:

```text
DATA|seq|total|payload_criptografado|checksum
```

Exemplo:

```text
DATA|0|8|1d1c0203|180
```

Campos:

- `seq`: numero de sequencia do fragmento;
- `total`: quantidade total de fragmentos da mensagem;
- `payload_criptografado`: fragmento criptografado em hexadecimal;
- `checksum`: soma dos caracteres do payload original, modulo 256.

### ACK

Enviado pelo servidor quando um pacote e aceito:

```text
ACK|seq
```

### NAK

Enviado pelo servidor quando um pacote precisa ser retransmitido:

```text
NAK|seq
```

O `seq` indica qual pacote deve ser reenviado ou qual sequencia o servidor esta esperando.

## Fragmentacao

O cliente divide a mensagem em blocos de 4 caracteres:

```go
fragSize := 4
```

Se o ultimo fragmento tiver menos de 4 caracteres, ele e preenchido com espacos antes de ser criptografado:

```go
payload := fmt.Sprintf("%-4s", fragment)
```

No servidor, esses espacos extras sao removidos antes da remontagem:

```go
payload := strings.TrimRight(payloadPad, " ")
```

## Criptografia Simples

O cliente aplica uma criptografia manual usando a chave:

```go
COMP
```

O processo faz:

1. XOR entre cada byte do fragmento e cada byte da chave;
2. troca de posicoes dos bytes;
3. conversao do resultado para hexadecimal.

O servidor realiza o processo inverso em `manualDecrypt`.

Essa criptografia e apenas didatica. Ela serve para demonstrar transformacao do payload antes do envio, nao para seguranca real.

## Checksum e Deteccao de Erro

O checksum soma os caracteres do payload e aplica modulo 256:

```go
return sum % 256
```

Quando o servidor recebe um pacote:

1. descriptografa o payload;
2. recalcula o checksum;
3. compara com o checksum recebido.

Se os valores forem diferentes, o servidor registra erro e responde:

```text
NAK|seq
```

## Simulacao de Perdas e Erros

O cliente simula falhas automaticamente antes de enviar cada pacote.

As probabilidades atuais sao:

```go
lossProbabilityPercent  = 20
errorProbabilityPercent = 15
```

Isso significa:

- 20% de chance de um pacote ser descartado pelo proprio cliente, simulando perda;
- 15% de chance de um pacote ser enviado com checksum corrompido, simulando erro.

Logs esperados no cliente:

```text
PERDA SIMULADA seq 2 (envio)
ERRO SIMULADO seq 4 (envio)
Timeout -> Go-Back-N retransmitindo a partir da base 2
Timeout -> Selective Repeat retransmitindo apenas pendentes
```

## Modos de Operacao

### Go-Back-N

No modo `gobackn`, o servidor so aceita o pacote com o numero de sequencia esperado.

Se chegar um pacote futuro, por exemplo `seq=4` quando o servidor espera `seq=2`, ele responde:

```text
NAK|2
```

No cliente, quando ocorre timeout ou `NAK`, a retransmissao volta para a base da janela.

Esse comportamento representa o Go-Back-N: quando ha falha, o cliente reenvia a partir do primeiro pacote ainda nao confirmado.

### Selective Repeat

No modo `selecionado`, o servidor aceita pacotes fora de ordem e os armazena.

O cliente controla quais sequencias ja receberam `ACK`. Em caso de timeout, ele retransmite apenas os pacotes pendentes dentro da janela, em vez de reenviar todos desde a base.

Esse comportamento representa o Selective Repeat: apenas os pacotes perdidos ou com erro precisam ser reenviados.

## Validacoes do Servidor

O servidor possui validacoes para evitar falhas durante a execucao:

- rejeita `HELLO` malformado;
- rejeita `DATA` malformado;
- rejeita `DATA` antes de `HELLO`;
- valida numeros de sequencia e total;
- valida payload criptografado;
- valida checksum;
- responde `NAK` quando precisa solicitar retransmissao.

## Como Executar

Abra dois terminais.

### Servidor

```bash
cd comunicacoes/Server
go run .
```

Saida esperada:

```text
[SERVIDOR] ouvindo em 0.0.0.0:10000
```

### Cliente

```bash
cd comunicacoes/Client
go run .
```

Fluxo esperado no cliente:

1. escolher o modo:

```text
gobackn
```

ou:

```text
selecionado
```

2. digitar uma mensagem;
3. acompanhar os logs de envio, perda, erro, `ACK`, `NAK` e retransmissao.

## Como Testar

No modulo do servidor:

```bash
cd comunicacoes/Server
go test ./...
```

No modulo do cliente:

```bash
cd comunicacoes/Client
go test ./...
```

## Resumo Final

A implementacao atual demonstra comunicacao confiavel construida sobre UDP. O cliente envia dados fragmentados, simula perdas e erros, e o servidor confirma ou rejeita pacotes com `ACK` e `NAK`.

O modo `gobackn` retransmite a partir da base da janela quando ocorre falha. O modo `selecionado` aceita pacotes fora de ordem e retransmite apenas os pendentes. Com isso, o projeto cobre os principais comportamentos esperados para simular processos de transmissao com erro, perda e recuperacao.
