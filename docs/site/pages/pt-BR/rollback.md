# Reverter migrações

Executa Down ou `down(db)`. Só desfaz o que o código implementa, sem recuperar automaticamente dados apagados.

## Manter a versão de destino

```sh
migrate down 202611030001 --dry-run
migrate down 202611030001
migrate status
```

Reverte apenas versões superiores. Com 1, 2, 3 aplicadas, `down 2` remove 3 e mantém 1 e 2.

## Reverter todas

```sh
migrate down --all --dry-run
migrate down --all
migrate status
```

Destino e `--all` são exclusivos. Após sucesso, os arquivos voltam a `pending`. Verifique todo o intervalo e as [transações](./script-transactions). Um script sem down bloqueia o intervalo; down vazio não desfaz nada. DDL e operações sem transação podem permanecer após falha.
