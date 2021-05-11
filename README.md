# Sendsync

Sendsync is a cli written in golang for interacting with the Sendgrid API. We use it mainly to `GET` and `POST` templates and versions.

## Installation

Run
```bash
go build
go install
```
This will create a binary executable.
## Usage

First, export your `SENDGRID_API_KEY`.

To fetch all templates:
```bash
sendsync get templates
```

To publish a specific template:
```bash
sendsync apply templates/name_of_the_template/template.json
```

## Questions
Have a question? Come find us in #platform
