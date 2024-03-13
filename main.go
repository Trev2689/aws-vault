package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "aws-vault",
		Short: "CLI tool for interacting with AWS services",
		Run: func(cmd *cobra.Command, args []string) {
			// Display usage information when no command is provided
			cmd.Usage()
		},
	}

	// Add commands to root command
	rootCmd.AddCommand(uploadCmd)
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(updateSecretCmd)
	rootCmd.AddCommand(createSecretCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Command to upload file to S3
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload file to S3 bucket",
	Run: func(cmd *cobra.Command, args []string) {
		// Check required input parameters
		bucketName, _ := cmd.Flags().GetString("bucket")
		filePath, _ := cmd.Flags().GetString("file")
		subdirectory, _ := cmd.Flags().GetString("subdirectory")

		if bucketName == "" || filePath == "" {
			fmt.Println("Please provide all required input parameters: --bucket and --file")
			os.Exit(1)
		}

		// Load AWS SDK configuration
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			fmt.Println("Error loading AWS SDK config:", err)
			os.Exit(1)
		}

		// Create S3 client with the config from above
		client := s3.NewFromConfig(cfg)

		// Read file contents
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Println("Error reading file:", err)
			os.Exit(1)
		}

		// Upload file to S3
		_, err = client.PutObject(context.Background(), &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(subdirectory + "/" + filePath),
			Body:   bytes.NewReader(data),
		})
		if err != nil {
			fmt.Println("Error uploading file to S3:", err)
			os.Exit(1)
		}

		fmt.Println("File uploaded to S3 successfully.")
	},
}

// Command to download file from S3
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download file from S3 bucket",
	Run: func(cmd *cobra.Command, args []string) {
		// Check required input parameters
		bucketName, _ := cmd.Flags().GetString("bucket")
		filePath, _ := cmd.Flags().GetString("file")
		subdirectory, _ := cmd.Flags().GetString("subdirectory")

		if bucketName == "" || filePath == "" {
			fmt.Println("Please provide all required input parameters: --bucket and --file")
			os.Exit(1)
		}

		// Load AWS SDK configuration
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			fmt.Println("Error loading AWS SDK config:", err)
			os.Exit(1)
		}

		// Create S3 client with the config from above
		client := s3.NewFromConfig(cfg)

		// Download file from S3
		resp, err := client.GetObject(context.Background(), &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(subdirectory + "/" + filePath),
		})
		if err != nil {
			fmt.Println("Error downloading file from S3:", err)
			os.Exit(1)
		}

		// Read file contents
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading file contents:", err)
			os.Exit(1)
		}

		// Write file contents to disk
		err = ioutil.WriteFile(filePath, data, 0644)
		if err != nil {
			fmt.Println("Error writing file to disk:", err)
			os.Exit(1)
		}

		fmt.Println("File downloaded from S3 successfully.")
	},
}

// Command to update secret in Secrets Manager
var updateSecretCmd = &cobra.Command{
	Use:   "update-secret",
	Short: "Update secret in Secrets Manager",
	Run: func(cmd *cobra.Command, args []string) {
		// Check required input parameters
		secretName, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		description, _ := cmd.Flags().GetString("description")
		jsonFilePath, _ := cmd.Flags().GetString("json-file")
		timeout, _ := cmd.Flags().GetDuration("timeout")
		updateFlag, _ := cmd.Flags().GetBool("update")

		if secretName == "" || region == "" || description == "" || jsonFilePath == "" {
			fmt.Println("Please provide all required input parameters: --name, --region, --description, and --json-file")
			os.Exit(1)
		}

		// Read secret value from a JSON file
		secretValue, err := readSecretFromJSON(jsonFilePath)
		if err != nil {
			fmt.Println("Error reading secret value from JSON file:", err)
			os.Exit(1)
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Load AWS SDK configuration with the specified region
		cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
		if err != nil {
			fmt.Println("Error loading AWS SDK config:", err)
			os.Exit(1)
		}

		// Create Secrets Manager client with the config from above
		client := secretsmanager.NewFromConfig(cfg)

		// Check if the secret we want to create already exists
		describeInput := &secretsmanager.DescribeSecretInput{
			SecretId: &secretName,
		}

		_, err = client.DescribeSecret(ctx, describeInput)
		if err != nil {
			// Check if the error message indicates that the secret doesn't exist
			if strings.Contains(err.Error(), "ResourceNotFoundException") {
				// Secret does not exist, so proceed with creation
				fmt.Println("Secret does not exist. Proceeding with creation...")
			} else {
				// Some other error occurred, best to exit
				fmt.Println("Error describing secret:", err)
				os.Exit(1)
			}
		} else {
			// Secret already exists, check if update flag is set
			if !updateFlag {
				fmt.Println("Secret already exists and update flag not set. Exiting.")
				return
			}
		}

		if updateFlag {
			// Update input for the UpdateSecret API operation
			updateInput := &secretsmanager.UpdateSecretInput{
				SecretId:     &secretName,
				Description:  &description,
				SecretString: &secretValue,
			}

			// Call UpdateSecret API operation with above input
			_, err = client.UpdateSecret(ctx, updateInput)
			if err != nil {
				fmt.Println("Error updating secret:", err)
				os.Exit(1)
			}

			fmt.Println("Successfully updated secret.")
			return
		}

		// Create input for the CreateSecret API operation
		createInput := &secretsmanager.CreateSecretInput{
			Name:         &secretName,
			Description:  &description,
			SecretString: &secretValue,
		}

		// Call CreateSecret API operation with above input
		createOutput, err := client.CreateSecret(ctx, createInput)
		if err != nil {
			fmt.Println("Error creating secret:", err)
			os.Exit(1)
		}

		// Print ARN of the created secret
		fmt.Println("Successfully created, Secret ARN:", *createOutput.ARN)
	},
}

// Command to create secret in Secrets Manager
var createSecretCmd = &cobra.Command{
	Use:   "create-secret",
	Short: "Create secret in Secrets Manager",
	Run: func(cmd *cobra.Command, args []string) {
		// Check required input parameters
		secretName, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		description, _ := cmd.Flags().GetString("description")
		jsonFilePath, _ := cmd.Flags().GetString("json-file")
		timeout, _ := cmd.Flags().GetDuration("timeout")

		if secretName == "" || region == "" || description == "" || jsonFilePath == "" {
			fmt.Println("Please provide all required input parameters: --name, --region, --description, and --json-file")
			os.Exit(1)
		}

		// Read secret value from a JSON file
		secretValue, err := readSecretFromJSON(jsonFilePath)
		if err != nil {
			fmt.Println("Error reading secret value from JSON file:", err)
			os.Exit(1)
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Load AWS SDK configuration with the specified region
		cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
		if err != nil {
			fmt.Println("Error loading AWS SDK config:", err)
			os.Exit(1)
		}

		// Create Secrets Manager client with the config from above
		client := secretsmanager.NewFromConfig(cfg)

		// Create input for the CreateSecret API operation
		createInput := &secretsmanager.CreateSecretInput{
			Name:         &secretName,
			Description:  &description,
			SecretString: &secretValue,
		}

		// Call CreateSecret API operation with above input
		createOutput, err := client.CreateSecret(ctx, createInput)
		if err != nil {
			fmt.Println("Error creating secret:", err)
			os.Exit(1)
		}

		// Print ARN of the created secret
		fmt.Println("Successfully created, Secret ARN:", *createOutput.ARN)
	},
}

func readSecretFromJSON(filePath string) (string, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func init() {
	uploadCmd.Flags().StringP("bucket", "b", "", "S3 bucket name")
	uploadCmd.Flags().StringP("file", "f", "", "Path to file to upload")
	uploadCmd.Flags().StringP("subdirectory", "s", "", "Subdirectory in S3 bucket")

	downloadCmd.Flags().StringP("bucket", "b", "", "S3 bucket name")
	downloadCmd.Flags().StringP("file", "f", "", "Path to file to download")
	downloadCmd.Flags().StringP("subdirectory", "s", "", "Subdirectory in S3 bucket")

	updateSecretCmd.Flags().StringP("name", "n", "", "Name of the secret")
	updateSecretCmd.Flags().StringP("region", "r", "", "AWS region")
	updateSecretCmd.Flags().StringP("description", "d", "", "Description of the secret")
	updateSecretCmd.Flags().StringP("json-file", "j", "", "Path to JSON file containing secret value")
	updateSecretCmd.Flags().DurationP("timeout", "t", 30*time.Second, "Timeout for the operation")
	updateSecretCmd.Flags().BoolP("update", "u", false, "Update secret if it already exists")

	createSecretCmd.Flags().StringP("name", "n", "", "Name of the secret")
	createSecretCmd.Flags().StringP("region", "r", "", "AWS region")
	createSecretCmd.Flags().StringP("description", "d", "", "Description of the secret")
	createSecretCmd.Flags().StringP("json-file", "j", "", "Path to JSON file containing secret value")
	createSecretCmd.Flags().DurationP("timeout", "t", 30*time.Second, "Timeout for the operation")
}
