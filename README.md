# AWS VAULT CLI

CLI tool for interacting with AWS services including uploading files to S3 buckets and managing secrets in AWS Secrets Manager.

## Usage

1. **Upload a file to S3 bucket:**

    ```bash
    ./aws-vault upload --bucket <bucket_name> --file <file_path>
    ```

    Replace `<bucket_name>` with the name of your S3 bucket and `<file_path>` with the path to the file you want to upload.

2. **Download a file from S3 bucket:**

    ```bash
    ./aws-vault download --bucket <bucket_name> --file <file_path>
    ```

    Replace `<bucket_name>` with the name of your S3 bucket and `<file_path>` with the path to the file you want to download.

3. **Create or update a secret in Secrets Manager:**

    ```bash
    ./aws-vault create-secret --name <secret_name> --region <aws_region> --description <description> --json-file <json_file_path>
    ```

    Replace `<secret_name>` with the name of the secret, `<aws_region>` with the AWS region, `<description>` with the description of the secret, and `<json_file_path>` with the path to the JSON file containing the secret value.

    To update an existing secret, use the `--update` flag.

    ```bash
    ./aws-vault update-secret --name <secret_name> --region <aws_region> --description <description> --json-file <json_file_path>
    ```

    Replace the parameters as described above.

## Build

To build the CLI tool, run:

```bash
go build -o aws-vault main.go
