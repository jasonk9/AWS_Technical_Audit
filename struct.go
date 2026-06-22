package main

import "github.com/aws/aws-sdk-go-v2/service/iam/types"

type accountInfo struct {
	AccountID   string
	AccountName string
	Profile     string
}

//type UserAccessReport struct {
//	UserName              string
//	InlinePolcies         string
//	AttachedPolicies      string
//	Groups                string
//	GroupInlinePolicies   string
//	GroupAttachedPolicies string
//}

type UserData struct {
	UserName string
	InLine   []string
	Attached []types.AttachedPolicy
	Groups   []GroupData
}

/*
func (t *test) Size() int {

	if len(t.Groups) >= len(t.InLine) && len(t.Groups) >= len(t.Attached) {
		return len(t.Groups)
	} else if len(t.Attached) >= len(t.InLine) && len(t.Attached) >= len(t.Groups) {
		return len(t.Attached)
	} else {
		return len(t.InLine)
	}
}*/

type GroupData struct {
	GroupName string
	Inline    []string
	Attached  []types.AttachedPolicy
}

/*
func (t *test2) Size() int {
	if len(t.Attached) >= len(t.Inline) {
		return len(t.Attached)
	} else {
		return len(t.Inline)
	}
}*/
