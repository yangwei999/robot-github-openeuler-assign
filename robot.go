package main

import (
	"fmt"
	sdk "github.com/google/go-github/v36/github"
	"github.com/opensourceways/community-robot-lib/config"
	gc "github.com/opensourceways/community-robot-lib/githubclient"
	framework "github.com/opensourceways/community-robot-lib/robot-github-framework"
	"github.com/sirupsen/logrus"
)

const (
	botName      = "issue-assign"
	createAction = "created"
	state        = "open"
)

type iClient interface {
	ListCollaborator(pr gc.PRInfo) ([]*sdk.User, error)
	AssignSingleIssue(is gc.PRInfo, login string) error
	UnAssignSingleIssue(is gc.PRInfo, login string) error
	CreateIssueComment(is gc.PRInfo, comment string) error
}

func newRobot(cli iClient) *robot {
	return &robot{cli: cli}
}

type robot struct {
	cli iClient
}

func (bot *robot) NewConfig() config.Config {
	return &configuration{}
}

func (bot *robot) getConfig(cfg config.Config, org, repo string) (*botConfig, error) {
	c, ok := cfg.(*configuration)
	if !ok {
		return nil, fmt.Errorf("can't convert to configuration")
	}

	if bc := c.configFor(org, repo); bc != nil {
		return bc, nil
	}

	return nil, fmt.Errorf("no config for this repo:%s/%s", org, repo)
}

func (bot *robot) RegisterEventHandler(f framework.HandlerRegister) {
	f.RegisterIssueCommentHandler(bot.handleIssueCommentEvent)
}

func (bot *robot) handleIssueCommentEvent(e *sdk.IssueCommentEvent, cfg config.Config, log *logrus.Entry) error {
	if e.GetAction() != createAction || e.GetIssue().GetState() != state {
		return nil
	}

	org, repo := gc.GetOrgRepo(e.GetRepo())
	c, err := bot.getConfig(cfg, org, repo)
	if err != nil {
		return err
	}

	if c == nil {
		return nil
	}

	return bot.handleAssign(e, org, repo)
}
