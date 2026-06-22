# Changelog

All notable changes will documented here

### [1.0.0] - 2026-06-22

Initial Release.  

- Fetch AWS account details from github repo
- Generates and downloads AWS credential report for each account in its own directory
- Generates a csv per IAM user in the accounts directory
  - Showing relationships for the IAM User to policies applied
- Downloads a copy of all policies attached to the IAM Uuser into the AWS Accounts directory
- Downloads a copy of all policies attached to groups the IAM User is a memberOf into the AWS Accounts directory