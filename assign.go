package main

import (
	"fmt"
	sdk "github.com/google/go-github/v36/github"
	gc "github.com/opensourceways/community-robot-lib/githubclient"
	"strings"
)

const (
	msgMultipleAssignee = "Can only assign one assignee to the issue."
	msgAssignRepeatedly = "This issue is already assigned to ***%s***. Please do not assign repeatedly."
	msgNotAllowAssign   = "This issue can not be assigned to ***%s***. Please try to assign to the repository collaborators."
	msgNotAllowUnassign = "***%s*** can not be unassigned from this issue. Please try to unassign the assignee of this issue."
)

func (bot *robot) handleAssign(e *sdk.IssueCommentEvent, org, repo string) error {
	is := gc.PRInfo{Org: org, Repo: repo, Number: e.GetIssue().GetNumber()}
	comment := e.GetComment().GetBody()
	commenter := e.GetComment().GetUser().GetLogin()

	currentAssignee := ""

	if e.GetIssue().GetAssignee() != nil {
		currentAssignee = e.GetIssue().GetAssignee().GetLogin()
	}

	writeComment := func(s string) error {
		return bot.cli.CreateIssueComment(is, s)
	}

	assign, unassign := parseCmd(comment, commenter)
	fmt.Println("as unas ", assign, unassign)
	if n := assign.Len(); n > 0 {
		if n > 1 {
			return writeComment(msgMultipleAssignee)
		}

		if assign.Has(currentAssignee) {
			return writeComment(fmt.Sprintf(msgAssignRepeatedly, currentAssignee))
		}

		newOne := assign.UnsortedList()[0]

		err := bot.cli.AssignSingleIssue(is, newOne)
		if err == nil {
			return nil
		}
		if err != nil {
			return writeComment(fmt.Sprintf(msgNotAllowAssign, newOne))
		}
		return err
	}

	if unassign.Len() > 0 {
		if unassign.Has(currentAssignee) {
			err := bot.cli.UnAssignSingleIssue(is, currentAssignee)
			fmt.Println(err)
			return err
		} else {
			return writeComment(fmt.Sprintf(msgNotAllowUnassign, strings.Join(unassign.UnsortedList(), ",")))
		}
	}

	return nil
}
