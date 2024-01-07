# drone

A tool to import IoC feeds from provider and save records to BigQuery table.

![logo](https://github.com/m-mizutani/drone/assets/605953/f1ba68ae-184c-4342-a32f-70663e865902)

## Features

* Import IoC feeds from provider, currently supporting
    * [AlienVault OTX](https://otx.alienvault.com/) (subscribed pulses)
    * [Abuse.ch](https://abuse.ch/) (Feodo)
* Prevent duplicated records by imported time

## Usage

### Prerequisite

* [Google Cloud Platform](https://cloud.google.com/) account
* [BigQuery](https://cloud.google.com/bigquery/) dataset
* Service Account with BigQuery write permission of the dataset such as `roles/bigquery.dataEditor`
* Service Account key file (JSON)
* Each providers account (if you need)
    * AlienVault OTX (API key)

### Installation

```bash
$ go install github.com/m-mizutani/drone@latest
```

or

```bash
$ docker run ghcr.io/m-mizutani/drone:latest
```

### Usage

#### Import AlienVault OTX pulses

```bash
$ export DRONE_BIGQUERY_PROJECT_ID=your-project-id
$ export DRONE_BIGQUERY_DATASET_ID=your_dataset_id
$ export DRONE_BIGQUERY_SA_KEY_FILE=/path/to/your_service_account_key.json
# If you want to set credential directly, use DRONE_BIGQUERY_SA_KEY_DATA
# export DRONE_BIGQUERY_SA_KEY_DATA=$(cat /path/to/your_service_account_key.json)
$ export DRONE_OTX_API_KEY=abcde12345XXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
$ drone import otx subscribed
```

#### Import Abuse.ch Feodo

```bash
$ export DRONE_BIGQUERY_PROJECT_ID=your-project-id
$ export DRONE_BIGQUERY_DATASET_ID=your_dataset_id
$ export DRONE_BIGQUERY_SA_KEY_FILE=/path/to/your_service_account_key.json
$ drone import abusech feodo
```

## License

Apache License 2.0
