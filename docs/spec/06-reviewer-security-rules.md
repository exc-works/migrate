# 06 Reviewer & Security Rules

## 阻断项（必须拒绝合并）

- 提交中出现绝对路径。
- 提交中出现明文密码、Token、访问凭据。
- 提交中出现私钥或密钥材料（例如 `BEGIN PRIVATE KEY`）。

## 配套规则

- 示例配置必须使用占位符。
- 日志和报错输出必须脱敏。
- 测试样例不得使用真实账号/真实连接串。

## Pre-commit 检查建议

在 `pre-commit` hook 中至少增加一类敏感信息扫描：

- 私钥特征匹配。
- 常见密码/Token 模式匹配。
- 高风险连接串字段匹配（如 `password=`、`secret=`、`token=`）。
