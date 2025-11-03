# pkg/database

该包为所有服务提供统一的数据库访问能力，聚焦以下内容：

- 连接池配置：`NewDatabase` / `NewDatabaseWithConfig` 负责应用推荐的连接参数（最大连接 25、空闲 5、定期回收）。
- 事务包装：`WithTx` 提供基于回调的事务统一入口，并返回抽象 `Transaction` 接口，方便测试与解耦。
- Outbox 仓储：通过 `NewOutboxRepository` 实现事务性发件箱的持久化，配合 Plan 217B 的 dispatcher 使用。
- 指标采集：`RegisterMetrics`、`RecordConnectionStats` 与查询包装方法提供 Prometheus 观测能力。

使用建议：

```go
dbClient, err := database.NewDatabase(os.Getenv("DATABASE_DSN"))
if err != nil {
    log.Fatal(err)
}

database.RegisterMetrics(prometheus.DefaultRegisterer)

repo := database.NewOutboxRepository(dbClient)

err = dbClient.WithTx(ctx, func(ctx context.Context, tx database.Transaction) error {
    // 业务数据写入
    if _, err := tx.ExecContext(ctx, "UPDATE organizations SET name = $1 WHERE code = $2", name, code); err != nil {
        return err
    }

    return repo.Save(ctx, tx, &database.OutboxEvent{
        EventID:       uuid.NewString(),
        AggregateID:   code,
        AggregateType: "organization",
        EventType:     "organization.updated",
        Payload:       payloadJSON,
    })
})
if err != nil {
    return err
}
```

周期性调用 `dbClient.RecordConnectionStats(serviceName)` 可同步连接池指标。
