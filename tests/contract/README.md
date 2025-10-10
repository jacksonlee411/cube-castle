# Contract Snapshot Tests

- 更新 `docs/api/openapi.yaml`、`docs/api/schema.graphql` 或 `shared/contracts/organization.json` 后执行 `scripts/contract/sync.sh` 并运行 `python tests/contract/verify_inventory.py`。
- 如预期发生差异，请更新 `tests/contract/inventory.baseline.json`。
