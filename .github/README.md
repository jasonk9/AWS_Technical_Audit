# Overview

Utility reads configuration files from github so no updates are required when AWS changes occur.  

For every AWS account defined in the account.json file in github this utility will:   
- Generate and Download the accounts Credential Report  
- For each IAM User account
  - Download and save inline policies
  - Download and save attached policies
  - Get groups the user is a memeber of
  - for each group
    - Download and save inline policies
    - Download and save attached policies
  - Wtite a summary table for the auditor

# Setup.  

## Github

1. Create a personal token.  
```
From github settings
Select deveolper settings
Select personnal access tokens
Select Fine-grained tokens
Select Only select repositories - AWSInfo
Permissions -
  Contents - read-only
  Metadata - read-only
  Pages - readonly
```  

## macOS.  

1. Create a new Keychain entry:   
```
    Chain:   login.   
    Name:    token_github_awsaudit.  
    Kind:    application password.   
    Account: anon.   
    Token:   your github token
```

## Windows.  

1. Create a new Credential Manager Entry:    
```
To Do need a windows peep to guinie pig
```

# Usage.  

## Default

./AWSAudit [enter]        

Default Values:   
|  Option. |. Value  |
| -------- | ------- |
| owner | Geoscape |
| repo | aws-info |
| accountsFile | accounts.json |
| token | token_github_awsaudit |

## Custom values

./AWSAudit -owner=example -repo=sample -accountsFile=myaccounts.json -tokenName=sample_token.   

## Testing

./AWSAudit -owner=jasonk9 -repo=readfiletest -accountsFile=accounts.json -tokenName=token_github_test