package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/zalando/go-keyring"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

// func LoadAccountsFromGithub() ([]accountInfo, error) {
func GetAccountsFromGithub(owner string, repo string, path string, token string) ([]accountInfo, error) {

	//
	//. Temporary Change this to be the real name
	//
	token, err := keyring.Get(token, "anon")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	//
	//. Setup auth to github
	//
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	opts := &github.RepositoryContentGetOptions{Ref: "main"}
	//fileContent, _, _, err := client.Repositories.GetRepositoryContent(ctx, owner, repo, path, opts)
	fileContent, _, _, err := client.Repositories.GetContents(context.Background(), owner, repo, path, opts)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	jsonRaw, err := fileContent.GetContent()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	var accounts []accountInfo
	err = json.Unmarshal([]byte(jsonRaw), &accounts)

	return accounts, nil
}

func GetAWSCredentialReportForAccount(account *accountInfo) {

	fullPath := filepath.Join(account.AccountID, "Credential-Report.csv")
	if !fileExists(fullPath) {
		cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile(account.Profile))
		if err != nil {
			log.Fatal(err)
		}

		client := iam.NewFromConfig(cfg)

		genOut, err := client.GenerateCredentialReport(context.Background(), &iam.GenerateCredentialReportInput{})
		if err != nil {
			log.Fatal(err)
		}

		for genOut.State != types.ReportStateTypeComplete {
			time.Sleep(3 * time.Second)

			genOut, err = client.GenerateCredentialReport(context.Background(), &iam.GenerateCredentialReportInput{})
			if err != nil {
				log.Fatal(err)
			}
		}

		getOut, err := client.GetCredentialReport(context.Background(), &iam.GetCredentialReportInput{})

		os.MkdirAll(account.AccountID, 0755)
		os.WriteFile(fullPath, getOut.Content, 0644)
	}

}

func GetInlineUserPolicy(account accountInfo, userName string, policyName string, config *aws.Config, client *iam.Client) {

	input := &iam.GetUserPolicyInput{
		UserName:   &userName,
		PolicyName: &policyName,
	}

	output, err := client.GetUserPolicy(context.TODO(), input)
	if err != nil {
		log.Fatal(err)
	}

	saveJSON(&account, output.PolicyDocument, policyName)
}

func GetInlineGroupPolicy(account accountInfo, groupName string, policyName string, config *aws.Config, client *iam.Client) {

	input := &iam.GetGroupPolicyInput{
		GroupName:  &groupName,
		PolicyName: &policyName,
	}

	output, err := client.GetGroupPolicy(context.TODO(), input)
	if err != nil {
		log.Fatal(err)
	}

	saveJSON(&account, output.PolicyDocument, policyName)
}

func GetAttachedPolicy(account accountInfo, policyARN *string, config *aws.Config, client *iam.Client) {
	result, err := client.GetPolicy(context.TODO(), &iam.GetPolicyInput{
		PolicyArn: policyARN,
	})
	if err != nil {
		log.Fatal(err)
	}

	policyResult, err := client.GetPolicyVersion(context.TODO(), &iam.GetPolicyVersionInput{
		PolicyArn: policyARN,
		VersionId: result.Policy.DefaultVersionId,
	})

	if err != nil {
		log.Fatal(err)
	}

	saveJSON(&account, policyResult.PolicyVersion.Document, *result.Policy.PolicyName)
}

func saveJSON(account *accountInfo, content *string, filename string) {
	fullPath := filepath.Join(account.AccountID, filename+".json")
	if !fileExists(fullPath) {

		decodedPolicy, err := url.QueryUnescape(aws.ToString(content))
		var prettyJSON map[string]interface{}
		err = json.Unmarshal([]byte(decodedPolicy), &prettyJSON)
		if err != nil {
			log.Fatal(err)
		}

		prettyPolicy, err := json.MarshalIndent(prettyJSON, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		os.WriteFile(fullPath, []byte(prettyPolicy), 0644)
	}
}

func fileExists(fileName string) bool {
	_, err := os.Stat(fileName)
	if err == nil {
		return true //file exists
	}
	if errors.Is(err, os.ErrNotExist) {
		return false //File does not exist
	}

	return false //File may or may not exist I.E. permission denied etc
}

func directoryExists(path string) bool {
	info, err := os.Stat(path)
	if err == nil {
		return info.IsDir()
	}

	if errors.Is(err, os.ErrNotExist) {
		return false
	}

	return false
}

func writeToCSV(userData UserData, account *accountInfo) error {
	filename := filepath.Join(account.AccountID, fmt.Sprintf("%s.csv", userData.UserName))

	if fileExists(filename) {
		return nil //skip for now
	}

	if !directoryExists(account.AccountID) {
		os.MkdirAll(account.AccountID, 0755)
	}

	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err) //this doent need to be fatal will do for now
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"Type", "Group Name", "Policy Type", "Policy Name"}
	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, policy := range userData.InLine {
		row := []string{"User", "", "InLine", policy}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	for _, policy := range userData.Attached {
		row := []string{"User", "", "Attached", *policy.PolicyName}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	for _, group := range userData.Groups {
		for _, policy := range group.Inline {
			row := []string{"Group", group.GroupName, "InLine", policy}
			if err := writer.Write(row); err != nil {
				return err
			}
		}

		for _, policy := range group.Attached {
			row := []string{"Group", group.GroupName, "Attached", *policy.PolicyName}
			if err := writer.Write(row); err != nil {
				return err
			}
		}
	}
	return nil
}
