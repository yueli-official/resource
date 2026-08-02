# User 主体合同消费者迁移

## Status

Finished

## Result

- Resource view/controller/authz 只按 verified `subject_kind` 接受 user/client，ping 回显声明的 user、guest 或 client actor。
- 本地 bootstrap 管理员改用 Identity Public User Key。
- 测试 principal 由真实 JWT 签发与验签构造，并修正旧 server test 的 verifier typo。

## Verification

- `go test ./... -count=1 -timeout 240s`
- `go vet ./...`

全局合同和发布顺序记录在 Identity User 合同 Work 中。
