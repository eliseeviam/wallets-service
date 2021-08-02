# Wallets service

Wallets service provides API for creating and fetching wallets, making deposits and transfers between wallets, fetching transfers history

### Running

Database schema in `./.postgres_init/schema.sql`. 
Schema apply on database while initial start. 
For recreating database `rm -r .postgres_data/ .redis_data/`

Run service with
```
$ docker-compose up wallet
[+] Running 4/4
 ⠿ Network wallets-service_default       Created                        
 ⠿ Container wallets-service_redis_1     Started                        
 ⠿ Container wallets-service_postgres_1  Started                        
 ⠿ Container wallets-service_wallet_1    Started
```

### Common response codes

1. `200` – success;
1. `201` – idempotency key duplication;
1. `400` – request parameters or idempotency key error;
1. `402` – insufficient balance in case of transfer;   
1. `404` – wallet not found;
1. `405` – request method mismatch;
1. `409` – wallet already created(creating with name of existing one);
1. `500` – internal error;

Error details provides with JSON

```json
{
  "ok": 0,
  "message": "reason"
}
```

### Create wallet

#### Request

```
$ curl --location --request POST 'localhost:8080/wallet' \
--header 'Idempotency-Key: XXXYYY' \
--header 'Content-Type: application/json' \
--data-raw '{
    "wallet_name": "myWallet"
}'
```

#### Response

In case of error, standart explanation

### Fetch wallet

#### Request

```
$ curl --location --request GET 'localhost:8080/wallet/myWallet'
```

#### Response

```json
{
    "name": "myWallet",
    "balance": 0
}
```

In case of error, standart explanation

### Deposit 

#### Request

```
$ curl --location --request POST 'localhost:8080/deposit' \
--header 'Idempotency-Key: XXXXYYYY' \
--header 'Content-Type: application/json' \
--data-raw '{
    "wallet_name": "myWallet",
    "amount": "1000"
}'
```

#### Response

In case of error, standart explanation

### Transfer 

#### Request

```
$ curl --location --request POST 'localhost:8080/transfer' \
--header 'Idempotency-Key: XXXXXYYYYY' \
--header 'Content-Type: application/json' \
--data-raw '{
    "wallet_name_from": "myWallet",
    "wallet_name_to": "anotherWallet",
    "amount": "500"
}'
```

#### Response

In case of error, standart explanation

### History 

#### Request

```
curl --location --request GET 'localhost:8080/history/myWallet'
```

##### Optional parameters
1. `direction` – deposit or transfer;
1. `start_date` – date in format `YYYY-MM-DD`(inclusively);
1. `end_date` – date in format `YYYY-MM-DD`(inclusively);
1. `limit`, `offset_by_id` - limit and offset can be used for pagination;

#### Response

```csv
id,amount,direction,meta,time
1,1000,deposit,"{""source"": ""unknown""}",2021-07-30T09:44:40.69095Z
1,-500,transfer,"{""destination"": ""anotherWallet""}",2021-07-30T09:46:37.625009Z
```

In case of error, standart explanation

### Configuration

Configuration provided from environment variables

|Name|Descrption|Constraint|
|---|---|---|
|`BIND_ADDR`| Address for binding | Not empty |
|`GRACEFUL_SHUTDOWN_TIMEOUT_SEC`| Waiting time before force service shutdown. -1 - unlimited | Greater or equal than -1 |
|`DB_TYPE`| RDBMS database type | `psql` for PostgreSQL |
|`DB_HOST`| Database host | Not empty |
|`DB_PORT`| Database port | Not empty |
|`DB_DATABASE_NAME`| Database name | Not empty |
|`DB_USER`| Database user | Not empty |
|`DB_PASSWORD`| Database password | Not empty |
|`IDEMPOTENCY_REDIS_ADDR`| Address of redis which set up for idempotency keys keeping | Any valid address representation |
|`IDEMPOTENCY_REDIS_PASSWORD`| Password of redis which set up for idempotency keys keeping | Any |
