These migrations are only necessary when upgrading from a previous version to a tag. If you are upgrading from
`0.0.1alpha2` to `0.0.1alpha4` in a running system you need to execute:

```
psql -h '' -U '' -d '' -f 0.0.1alpha3.sql  # 0.0.1alpha2 -> 0.0.1alpha3
psql -h '' -U '' -d '' -f 0.0.1alpha4.sql  # 0.0.1alpha3 -> 0.0.1alpha4
```

For a fresh install use `reset_db.sql`