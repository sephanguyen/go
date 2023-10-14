# Generate **ROW LEVEL SECURITY**

Granted View. **Required**. Make sure the view already existed in database.
```
go run cmd/utils/main.go pg_gen_rls  --rlsType=view --databaseName=bob

```
## **Generate Access Control from file**
---
For easy to use, track the work from **Access Control**. We can use **template** as a file to generate Access Control for both **PostgreSQL** and **Hasura** instead of use **CLI**.

### **Folder and file structure:**

```
backend
├─accesscontrol
|   ├─ {service-name}
|   |   ├─ {team}-{table}.yaml
|   |   ├─ {team}-{table}.yaml
|   |   ├─ {team}-{table}.yaml
|   ├─ bob
|   |   ├─ user-students.yaml
|   |   ├─ user-staffs.yaml
|   |   ├─ user-users.yaml
```


Main folder: **accesscontrol**

Each service will have children folder like: **accesscontrol/{service-name}**

**{service-name}** will have name like children folders of **migration** folder
Example: **bob**, **mastermgmt**, **...**

File name should be: **{team}-{table}.yaml**. **{team}** is your team name. **{table}** is table name which you want to apply Access Control.

### **Template structure**
```yaml
- template: {template-version}
  tableName: {main-table-name}
  locationCol: {location-column-name}
  accessPathTable:
    name: {access-path-table-name}
    columnMapping:
      {main-table-key}: {access-path-table-reference-key}
  permissionPrefix: {permission prefix}
  permissions:
    {type-of-system}: []
  ownerCol: {owner-column-name} 
  postgresPolicyVersion: {postgres-policy-version}
```
> **main table** mean the table we want to apply Access Control. **access path table** is table contain location of **main table**. if we save location inside **mai table** we can ignore **access path table**
>
>
>- **{template-version}**: Contain **1** or **1.1** or **4**. default is **1**
>- **{main-table-name}**: name of **main table**
>- **{location-column-name}**: name of location column. default is **location_id**
>- **{access-path-table-name}**: name of **access path table**
>- **{main-table-key}**: name of column primary key of **main table**
>- **{access-path-table-reference-key}** name of column have relationship with **main table** (usually have relationship with **{main-table-key}**) 
>- **{permission prefix}**: the prefix of permission. Example: we have permission user.students.read and user.student.write the prefix should be **user.students.
>- **{owner-column-name}**: the column of created record users. use for template 4 only.
>- **{type-of-system}**: Contain **postgres** and **hasura**. Which system you want to apply Access Control to.
>- **{postgres-policy-version}**: Option use in case we want to upgrade policy version without change any content on template

### **Run Template**
For all file
```sh
go run cmd/utils/main.go pg_gen_rls  --rlsType=rls-file
```
For all file on folder
```sh
go run cmd/utils/main.go pg_gen_rls  --rlsType=rls-file --acFolder=bob 
```
The behaviour will be the same with running by cli gen by PostgreSQL and Hasura. So For testing we can use CLI instead. 

Note: **At first version, we don't support stage for file yaml. So every change template need to be clear manually and re generate again. We will enhance in next version of template**

**Run template for stg**
```sh
go run cmd/utils/main.go pg_gen_rls  --rlsType=rls-file --stgHasura=true
```
**Run template for stg and specific database**
```sh
go run cmd/utils/main.go pg_gen_rls  --rlsType=rls-file --stgHasura=true --acFolder=bob
```
**Rollback command for one database**
```
go run cmd/utils/main.go pg_gen_rls  --rlsType=rollback-rls-file --acFolder=bob
```
**Rollback command for all database**
```
go run cmd/utils/main.go pg_gen_rls  --rlsType=rollback-rls-file
```

### **Testing**
For make sure our Access Control can be apply into all environment with out missing. We will base on the this template file to scan on Database policy and Hasura metadata for checking correctness. So for main applying we recommend using this template.


---

**Example file content:**

**Template 1:**

*Purpose use for team want to verify both location and permission of users*
```yaml
- template: 1
  tableName: lessons
  permissionPrefix: lesson.lessons
  permissions:
    postgres: []
    hasura: []
```

**Template 1.1:**

*Purpose use for team want to verify both location and permission of users without **insert** permission*

```yaml
- template: 1.1
  tableName: students
  locationCol: location_id
  accessPathTable:
    name: user_access_paths
    columnMapping:
      student_id: user_id
  permissionPrefix: user.student
  permissions:
    postgres: []
    hasura: []
```

**The template allow for <span style="color:red">INSERT</span> event we don't have permission or location.**

Most of case we should use **template 1.1** is location column not inside **main table**. And **main table** have foreign key with **access path table**. So the **INSERT** step can't work as well.

For using this template we recommend use template 1.1 for **main table** and template 1 for **access path table** and we insert data we cover 2 command insert to **main table** and **access path table** by one transaction.

Example:
```yaml
- template: 1.1
  tableName: students
  locationCol: location_id
  accessPathTable:
    name: user_access_paths
    columnMapping:
      student_id: user_id
  permissionPrefix: user.student
  permissions:
    postgres: []
    hasura: []
- template: 1
  tableName: user_access_paths
  permissionPrefix: user.student
  permissions:
    postgres: []
    hasura: []
```


**Template 4:**

*Purpose use for team want to verify owner of record only*
```yaml
- template: 4
  tableName: students
  ownerCol: owners 
  permissions:
    postgres: []
    hasura: []
```
Template 1 and 4:
**Purpose use for team want to verify location, permission and owner of each record**
```yaml
- template: 1.1
  tableName: students
  locationCol: location_id
  accessPathTable:
    name: user_access_paths
    columnMapping:
      student_id: user_id
  permissionPrefix: user.student
  permissions:
    postgres: []
    hasura: []
- template: 4
  tableName: students
  ownerCol: owners 
  permissions:
    postgres: []
    hasura: []
```


## **CLI**
---
### **Gen Granted Permission View**
This is required for all step below

```
go run cmd/utils/main.go pg_gen_rls  --rlsType=view --databaseName=bob 
```


### **PostgreSQL**
Example for **students** tables for 

```
go run cmd/utils/main.go pg_gen_rls  --rlsType=pg --table=students --pkey=student_id --accessPathTable=user_access_paths --databaseName=bob --permissionPrefix=user.student --accessPathTableKey=user_id
```
Example for **students** tables with template 4 owners check.
```
go run cmd/utils/main.go pg_gen_rls  --rlsType=pg --table=students --pkey=owners --templateVersion=4 --databaseName=bob
```
### **Hasura**

Example for **students** table
```
go run cmd/utils/main.go pg_gen_rls  --rlsType=hasura --table=students --pkey=student_id --accessPathTable=user_access_paths --databaseName=bob --permissionPrefix=user.student
```

Example for **students** table with Hasura V2
```
go run cmd/utils/main.go pg_gen_rls  --rlsType=hasura --table=students --pkey=student_id --accessPathTable=user_access_paths --databaseName=bob --permissionPrefix=user.student --hasuraVersion=2
```

Example for **students** table with Hasura V1 and template 4. include owner permission
```
go run cmd/utils/main.go pg_gen_rls  --rlsType=hasura --table=students --pkey=student_id --accessPathTable=user_access_paths --databaseName=bob --permissionPrefix=user.student --templateVersion=4 --ownerCol=owners

```
### **Gen Role MANABIE**
Command will help to generate role called MANABIE for Hasura. By combine all column and filters of other roles

Example generate role for database bob

```
go run cmd/utils/main.go pg_gen_rls  --rlsType=gen-role --databaseName=bob 
```

### **Options**
- **rlsType**: type included pg/hasura/view. pg is PostgreSQL RSL. hasura is Hasura RLS. view is generate migration file for granted_locations view.
- **table**: table which we want to generate RLS for
- **pkey**: primary key of table which related with access_path table or column content location. if template 4 of postgres this variable should be owner column name
- **accessPathTable**: Optional. Table which included access_path of input **table**
- **databaseName**: database name for PostgreSQL/View. service name for Hasura generation
- **permissionPrefix**: permission prefix. Example user.student
- **accessPathTableKey**: Optional. Column reference from main table to <table>_access_paths may have difference name with **pkey** column.
- **hasuraVersion**: Optional. Only need in case we generate code for hasura version 2. value: 1,2. default 1. if not existed this option hasura gen should be Hasura version 1
- **templateVersion**: Optional. Only need in case we generate code for template 4. value: 1,4. default 1. if not existed this option will generate template 1
- **ownerCol**: Required if templateVersion is 4 and rlsType is pg. Only need in case we generate code for template version 4. value: column which saved owners of record.
- **accessPathLocationCol**: in case location col in access path table is not location_id we can fill this option.


