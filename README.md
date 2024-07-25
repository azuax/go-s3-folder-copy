# GO S3 Folder Copy

Script for a fast folder upload to a desired AWS S3 bucket.
Useful if you want to backup your files on S3.

## Prerequisites
- Go installed
- Install dependencies
- Already created S3 bucket (Limit the permissions!)
- Set AWS credentials as environmental variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`. Ff needed: `AWS_SESSION_TOKEN` and `AWS_REGION`)


## Execution
```bash
go run main.go --path <path of the folder to upload> --bucket <Bucket name> [--base-folder-s3 <Prefix where to upload the objects>] [-t <number of workers>]
```

# ToDo
- [ ] Generate a summary of the uploaded files and errors


# Wishlist
- [ ] Handle file updates for faster sync and cron processes
- [ ] Add a README.md file in the base folder with information about the process.