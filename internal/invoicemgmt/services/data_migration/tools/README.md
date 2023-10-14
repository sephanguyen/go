## Data Migration Tools

These tools are Golang scripts that are used for data migration

### Invoicemgmt CSV generator script (generate_invoicemgmt_csv)

- This script generates invoice and payment CSV from the raw invoice data

**_NOTE:_**

- Before running this script, change the values returned by `getRawDataPath` and `generatedCSVDir` function with the path and directory that you want to use in your local.
- You can also change the maximum records per file by changing the value of `maxRowPerFile` in main.go
- If you didn't change the `generatedCSVDir`, it will create a `generated_csv` folder. This folder will contain a sub-folder that has different versions of the generated CSV.

```
generated_csv
└───version-20230316-154746
│   └───invoice
│       │   invoice_1.csv
│       │   invoice_2.csv
│       │   ...
│   └───payment
│       │   payment_1.csv
│       │   payment_2.csv
│       │   ...

```

#### How to run the generate_invoicemgmt_csv script?

- To run the script, use the below command

```bash
go run internal/invoicemgmt/services/data_migration/tools/generate_invoicemgmt_csv/cmd/main.go
```



### Invoicemgmt CSV generator Student Mapping script (generate_student_map)

- This script generates new invoice and payment CSV with mapped student id from external id
- To run the script, use the below command

**_NOTE:_** 
- Before running this script, change the value returned by `useHomeDir` function with the path and directory that you want to use in your local.
- Setup folder structure with csvs to be mapped from the path and directory you used
    - invoice_csv
      - mapped_invoice_csv
      - test_invoice.csv
    - payment_csv
      - mapped_payment_csv
      - test_payment.csv
    - user_mapping_id.csv
     
- Additional note on folder structure
   - user_mapping_id.csv naming is fixed
   - you can add multiple invoice or payment files regardless of naming under invoice_csv or payment_csv
   - example:
      - invoice_csv
        - test_invoice1.csv
        - test_1.csv
   - new csvs will be mapped under <entity>_csv/mapped_<entity>_csv
     - example **new_test_invoice1.csv** is generated: 
        - invoice_csv
          - test_invoice1.csv 
          - mapped_invoice_csv
            - new_test_invoice1.csv
        - user_mapping_id.csv
```bash
go run internal/invoicemgmt/services/data_migration/tools/generate_student_map_csv/cmd/main.go INVOICE_ENTITY
go run internal/invoicemgmt/services/data_migration/tools/generate_student_map_csv/cmd/main.go PAYMENT_ENTITY
```
