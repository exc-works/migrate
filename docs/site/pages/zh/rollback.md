# 回退迁移

回退执行迁移提供的 Down SQL 或 `down(db)`。
它只能撤销这些代码实际实现的操作，无法自行恢复已经删除的数据。

## 保留目标版本

```sh
migrate down 202611030001 --dry-run
migrate down 202611030001
migrate status
```

只有高于目标的已应用版本会被回退，目标版本保持已应用。
例如 1、2、3 均已应用时，`migrate down 2` 只回退版本 3，保留版本 1 和 2。

## 回退全部已应用版本

```sh
migrate down --all --dry-run
migrate down --all
migrate status
```

目标版本和 `--all` 互斥。成功回退后，对应文件恢复为待执行状态，之后可以再次应用。

## 检查整个回退范围

检查所有选中迁移的 Down 实现，以及数据库的事务限制。
脚本回退会在写入前准备整个选中范围；缺少 `down(db)` 会阻断该范围。
空的 down 函数并不能撤销 up 的变更。

DDL 隐式提交与非事务变更可能在回退失败后仍然保留。
恢复前应先检查数据和历史，参见[事务与执行限制](./script-transactions)及[问题排查](./troubleshooting)。
