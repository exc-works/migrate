# Aplicar migrações

Confira a conexão de destino e os arquivos.

```sh
migrate create
```

Cria apenas o histórico, se necessário. Se as mudanças já existem, veja [baseline](./history#baseline).

```sh
migrate up --dry-run
migrate up
migrate status
```

Aplica pendentes em ordem e registra os sucessos. Repetir não executa versões aplicadas. Uma execução silenciosa pode ser correta; confira `status`. Dry-run pode conectar e abrir SSH para ler o histórico, mas não aplica mudanças. Verifica scripts sem executá-los e não prevê SQL dinâmico.

Versões pendentes abaixo das aplicadas são ignoradas por padrão. Ative `migrate_out_of_order` apenas após verificar a compatibilidade com o esquema atual.

## Deploy

```sh
migrate --non-interactive up
migrate --non-interactive status --output json
```

Forneça variáveis e segredos antecipadamente e serialize comandos no mesmo banco. O histórico não garante execução única perante concorrência, quedas ou commits implícitos. Após falhar, inspecione o estado antes de repetir. Veja [problemas](./troubleshooting).
