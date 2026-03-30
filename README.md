
# Projeto de comunicacoes

## Execucao

### Va ao diretorio do Servidor e inicie o servidor WebSocket

```
  cd Server 
  go run .
```

O servidor expõe o endpoint `ws://localhost:8080/ws` e exige um handshake JSON inicial com:

- `operation_mode`
- `max_message_size`

`operation_mode` agora é escolhido pelo cliente e deve ser `gbn` (`Go-Back-N`) ou `sr` (`Selective Repeat`).

### Va ao ao diretorio do Cliente e inicie a conexao

```
  cd Client 
  go run .
```

### Envie a moeda desejada
