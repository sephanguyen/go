Run Data Warehouse with Kafka Access Control:
```
DATA_WH=true DATA_WH_AC=true ./deployments/sk.bash
```

Run Data Warehouse without Kafka Access Control:
```
DATA_WH=true ./deployments/sk.bash
```
Run Data Warehouse with specific migration file:
```
DATA_WH_FILE_FILTER="0007" DATA_WH=true ./deployments/sk.bash
```

Run hephaestus only for migration ksql DWH:
```
DATA_WH=true ./deployments/sk.bash -- -f skaffold2.backend.yaml -p hephaestus-only
```
Note: only use after run `DATA_WH=true ./deployments/sk.bash`

Run hephaestus with specific file migrations:
```
DATA_WH=true DATA_WH_FILE_FILTER="0007" build -- -f skaffold2.backend.yaml -p hephaestus-only
```

### If environment DATA_WH not set:
- Upsert kafka connector of Data Warehouse will skip
- Migrations of Data Lake will skip
- Migrations of Data Warehouse will skip
- Migrations of KSQL Data Warehouse will skip

### If environment DATA_WH_AC is true:
- Deploy Kafka of Data Warehouse with SASL PlAIN TEXT
- Deploy Kafka Connect/CP Registry/Kafka Exporter will implement SASL PLAIN TEXT to call Kafka
