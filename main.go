package main

import (
	"context"
	"flag"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

func main() {

	owner := flag.String("owner", "geoscape", "The owner of the github repo")
	repo := flag.String("repo", "aws-info", "The repo tto access")
	accountsFile := flag.String("accountsFile", "accounts.json", "File that holds the AWS account data")
	tokenName := flag.String("tokenName", "token_github_awsaudit", "Token to fetch for github auth")
	flag.Parse()

	accounts, err := GetAccountsFromGithub(*owner, *repo, *accountsFile, *tokenName)
	//accounts, err := LoadAccountsFromGithub()
	if err != nil {
		log.Fatal(err)
	}

	//
	//.  Get the AWS Credential report for each account and save it as a csv
	//
	if len(accounts) > 0 {
		//ACCOUNTS:
		for _, account := range accounts {
			//fmt.Println("ID: ", account.AccountID, "Name: ", account.AccountName, "Profile: ", account.Profile)
			var report UserData
			//Genertae and Save the accounts Credential Report
			go GetAWSCredentialReportForAccount(&account)

			//
			//.  Get the user list from AWS API
			//.  Simpler and quicker than extracting it from the report
			//.  Also dont need to wait for the report to generate
			//
			cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile(account.Profile))
			if err != nil {
				log.Fatal(err)
			}
			client := iam.NewFromConfig(cfg)

			userList, err := client.ListUsers(context.TODO(), &iam.ListUsersInput{})

			// for each user get policies and groups from aws
			if len(userList.Users) > 0 {

				//USERS:
				for _, user := range userList.Users {

					//Add the username to the report
					report.UserName = *user.UserName

					inLinePolicies, err := client.ListUserPolicies(context.TODO(), &iam.ListUserPoliciesInput{
						UserName: user.UserName,
					})
					if err != nil {
						log.Fatal(err)
					}

					if len(inLinePolicies.PolicyNames) > 0 {
						for _, inPolicy := range inLinePolicies.PolicyNames {
							go GetInlineUserPolicy(account, *user.UserName, inPolicy, &cfg, client)
						}
					}

					//Add inline policies to report
					report.InLine = inLinePolicies.PolicyNames

					attachedPolicies, err := client.ListAttachedUserPolicies(context.TODO(), &iam.ListAttachedUserPoliciesInput{
						UserName: user.UserName,
					})
					if err != nil {
						log.Fatal(err)
					}

					if len(attachedPolicies.AttachedPolicies) > 0 {
						for _, attPolicy := range attachedPolicies.AttachedPolicies {
							go GetAttachedPolicy(account, attPolicy.PolicyArn, &cfg, client)
						}
					}

					//Add inline policies to report
					report.Attached = attachedPolicies.AttachedPolicies

					groupMemberships, err := client.ListGroupsForUser(context.TODO(), &iam.ListGroupsForUserInput{
						UserName: user.UserName,
					})
					if err != nil {
						log.Fatal(err)
					}

					if len(groupMemberships.Groups) > 0 {
						//GROUPS:
						var g GroupData
						for _, group := range groupMemberships.Groups {
							g.GroupName = *group.GroupName

							groupInline, err := client.ListGroupPolicies(context.TODO(), &iam.ListGroupPoliciesInput{
								GroupName: group.GroupName,
							})
							if err != nil {
								log.Fatal(err)
							}

							if len(groupInline.PolicyNames) > 0 {
								g.Inline = groupInline.PolicyNames
								for _, gPolicy := range groupInline.PolicyNames {
									go GetInlineGroupPolicy(account, *group.GroupName, gPolicy, &cfg, client)
								}
							}

							groupAttached, err := client.ListAttachedGroupPolicies(context.TODO(), &iam.ListAttachedGroupPoliciesInput{
								GroupName: group.GroupName,
							})
							if err != nil {
								log.Fatal(err)
							}

							if len(groupAttached.AttachedPolicies) > 0 {
								g.Attached = groupAttached.AttachedPolicies
								for _, attPolicy := range groupAttached.AttachedPolicies {
									go GetAttachedPolicy(account, attPolicy.PolicyArn, &cfg, client)
								}
							}
							report.Groups = append(report.Groups, g)
						}
					}
					writeToCSV(report, &account)
					report = UserData{}
				}

			}
		}
	}
}
