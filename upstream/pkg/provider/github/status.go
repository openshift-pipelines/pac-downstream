package github

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v71/github"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/action"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/apis/pipelinesascode/keys"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/kubeinteraction"
	kstatus "github.com/openshift-pipelines/pipelines-as-code/pkg/kubeinteraction/status"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/opscomments"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/params/info"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/params/triggertype"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/provider"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
)

const (
	botType         = "Bot"
	pendingApproval = "Pending approval, waiting for an /ok-to-test"
)

const taskStatusTemplate = `
<table>
  <tr><th>Status</th><th>Duration</th><th>Name</th></tr>

{{- range $taskrun := .TaskRunList }}
<tr>
<td>{{ formatCondition $taskrun.PipelineRunTaskRunStatus.Status.Conditions }}</td>
<td>{{ formatDuration $taskrun.PipelineRunTaskRunStatus.Status.StartTime $taskrun.PipelineRunTaskRunStatus.Status.CompletionTime }}</td><td>

{{ $taskrun.ConsoleLogURL }}

</td></tr>
{{- end }}
</table>`

func (v *Provider) getExistingCheckRunID(ctx context.Context, runevent *info.Event, status provider.StatusOpts) (*int64, error) {
	opt := github.ListOptions{PerPage: v.PaginedNumber}
	for {
		res, resp, err := v.Client().Checks.ListCheckRunsForRef(ctx, runevent.Organization, runevent.Repository,
			runevent.SHA, &github.ListCheckRunsOptions{
				AppID:       v.ApplicationID,
				ListOptions: opt,
			})
		if err != nil {
			return nil, err
		}

		for _, checkrun := range res.CheckRuns {
			// if it is a Pending approval CheckRun then overwrite it
			if isPendingApprovalCheckrun(checkrun) || isFailedCheckrun(checkrun) {
				if v.canIUseCheckrunID(checkrun.ID) {
					return checkrun.ID, nil
				}
			}
			if *checkrun.ExternalID == status.PipelineRunName {
				return checkrun.ID, nil
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return nil, nil
}

func isPendingApprovalCheckrun(run *github.CheckRun) bool {
	if run == nil || run.Output == nil {
		return false
	}
	if run.Output.Title != nil && strings.Contains(*run.Output.Title, "Pending") &&
		run.Output.Summary != nil &&
		strings.Contains(*run.Output.Summary, "is waiting for approval") {
		return true
	}
	return false
}

func isFailedCheckrun(run *github.CheckRun) bool {
	if run == nil || run.Output == nil {
		return false
	}
	if run.Output.Title != nil && strings.Contains(*run.Output.Title, "pipelinerun start failure") &&
		run.Output.Summary != nil &&
		strings.Contains(*run.Output.Summary, "failed") {
		return true
	}
	return false
}

func (v *Provider) canIUseCheckrunID(checkrunid *int64) bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if v.checkRunID == 0 {
		v.checkRunID = *checkrunid
		return true
	}
	return false
}

func (v *Provider) createCheckRunStatus(ctx context.Context, runevent *info.Event, status provider.StatusOpts) (*int64, error) {
	now := github.Timestamp{Time: time.Now()}
	checkrunoption := github.CreateCheckRunOptions{
		Name:       provider.GetCheckName(status, v.pacInfo),
		HeadSHA:    runevent.SHA,
		Status:     github.Ptr("in_progress"),
		DetailsURL: github.Ptr(status.DetailsURL),
		ExternalID: github.Ptr(status.PipelineRunName),
		StartedAt:  &now,
	}

	checkRun, _, err := v.Client().Checks.CreateCheckRun(ctx, runevent.Organization, runevent.Repository, checkrunoption)
	if err != nil {
		return nil, err
	}
	return checkRun.ID, nil
}

func (v *Provider) getFailuresMessageAsAnnotations(ctx context.Context, pr *tektonv1.PipelineRun, pacopts *info.PacOpts) []*github.CheckRunAnnotation {
	annotations := []*github.CheckRunAnnotation{}
	r, err := regexp.Compile(pacopts.ErrorDetectionSimpleRegexp)
	if err != nil {
		v.Logger.Errorf("invalid regexp for filtering failure messages: %v", pacopts.ErrorDetectionSimpleRegexp)
		return annotations
	}
	intf, err := kubeinteraction.NewKubernetesInteraction(v.Run)
	if err != nil {
		v.Logger.Errorf("failed to create kubeinteraction: %v", err)
		return annotations
	}
	taskinfos := kstatus.CollectFailedTasksLogSnippet(ctx, v.Run, intf, pr, int64(pacopts.ErrorDetectionNumberOfLines))
	for _, taskinfo := range taskinfos {
		for _, errline := range strings.Split(taskinfo.LogSnippet, "\n") {
			results := map[string]string{}
			if !r.MatchString(errline) {
				continue
			}
			matches := r.FindStringSubmatch(errline)
			for i, name := range r.SubexpNames() {
				if i != 0 && name != "" {
					results[name] = matches[i]
				}
			}

			// check if we  have file in results
			var linenumber, errmsg, filename string
			var ok bool

			if filename, ok = results["filename"]; !ok {
				v.Logger.Errorf("regexp for filtering failure messages does not contain a filename regexp group: %v", pacopts.ErrorDetectionSimpleRegexp)
				continue
			}
			// remove ./ cause it would bug github otherwise
			filename = strings.TrimPrefix(filename, "./")

			if linenumber, ok = results["line"]; !ok {
				v.Logger.Errorf("regexp for filtering failure messages does not contain a line regexp group: %v", pacopts.ErrorDetectionSimpleRegexp)
				continue
			}

			if errmsg, ok = results["error"]; !ok {
				v.Logger.Errorf("regexp for filtering failure messages does not contain a error regexp group: %v", pacopts.ErrorDetectionSimpleRegexp)
				continue
			}

			ilinenumber, err := strconv.Atoi(linenumber)
			if err != nil {
				// can't do much regexp has probably failed to detect
				v.Logger.Errorf("cannot convert %s as integer: %v", linenumber, err)
				continue
			}
			annotations = append(annotations, &github.CheckRunAnnotation{
				Path:            github.Ptr(filename),
				StartLine:       github.Ptr(ilinenumber),
				EndLine:         github.Ptr(ilinenumber),
				AnnotationLevel: github.Ptr("failure"),
				Message:         github.Ptr(errmsg),
			})
		}
	}
	return annotations
}

// getOrUpdateCheckRunStatus create a status via the checkRun API, which is only
// available with GitHub apps tokens.
func (v *Provider) getOrUpdateCheckRunStatus(ctx context.Context, runevent *info.Event, statusOpts provider.StatusOpts) error {
	var err error
	var checkRunID *int64
	var found bool
	pacopts := v.pacInfo

	// The purpose of this condition is to limit the generation of checkrun IDs
	// when multiple pipelineruns fail. In such cases, generate only one checkrun ID,
	// regardless of the number of failed pipelineruns.
	if statusOpts.Title == "Failed" && statusOpts.PipelineRunName == "" {
		// setting different title to handle multiple checkrun cases
		statusOpts.Title = "pipelinerun start failure"
		if statusOpts.InstanceCountForCheckRun >= 1 {
			return nil
		}
	}

	// check if pipelineRun has the label with checkRun-id
	if statusOpts.PipelineRun != nil {
		var id string
		id, found = statusOpts.PipelineRun.GetAnnotations()[keys.CheckRunID]
		if found {
			checkID, err := strconv.Atoi(id)
			if err != nil {
				return fmt.Errorf("api error: cannot convert checkrunid")
			}
			checkRunID = github.Ptr(int64(checkID))
		}
	}
	if !found {
		if checkRunID, _ = v.getExistingCheckRunID(ctx, runevent, statusOpts); checkRunID == nil {
			checkRunID, err = v.createCheckRunStatus(ctx, runevent, statusOpts)
			if err != nil {
				return err
			}
		}
		if statusOpts.PipelineRun != nil {
			if _, err := action.PatchPipelineRun(ctx, v.Logger, "checkRunID and logURL", v.Run.Clients.Tekton, statusOpts.PipelineRun, metadataPatch(checkRunID, statusOpts.DetailsURL)); err != nil {
				return err
			}
		}
	}

	text := statusOpts.Text
	checkRunOutput := &github.CheckRunOutput{
		Title:   &statusOpts.Title,
		Summary: &statusOpts.Summary,
	}

	if statusOpts.PipelineRun != nil {
		if pacopts.ErrorDetection {
			checkRunOutput.Annotations = v.getFailuresMessageAsAnnotations(ctx, statusOpts.PipelineRun, pacopts)
		}
	}

	checkRunOutput.Text = github.Ptr(text)

	opts := github.UpdateCheckRunOptions{
		Name:   provider.GetCheckName(statusOpts, pacopts),
		Status: github.Ptr(statusOpts.Status),
		Output: checkRunOutput,
	}
	if statusOpts.PipelineRunName != "" {
		opts.ExternalID = github.Ptr(statusOpts.PipelineRunName)
	}
	if statusOpts.DetailsURL != "" {
		opts.DetailsURL = &statusOpts.DetailsURL
	}

	// Only set completed-at if conclusion is set (which means finished)
	if statusOpts.Conclusion != "" && statusOpts.Conclusion != "pending" {
		opts.CompletedAt = &github.Timestamp{Time: time.Now()}
		opts.Conclusion = &statusOpts.Conclusion
	}
	if isPipelineRunCancelledOrStopped(statusOpts.PipelineRun) {
		opts.Conclusion = github.Ptr("cancelled")
	}

	_, _, err = v.Client().Checks.UpdateCheckRun(ctx, runevent.Organization, runevent.Repository, *checkRunID, opts)
	return err
}

func isPipelineRunCancelledOrStopped(run *tektonv1.PipelineRun) bool {
	if run == nil {
		return false
	}
	if run.IsCancelled() || run.IsGracefullyCancelled() || run.IsGracefullyStopped() {
		return true
	}
	return false
}

func metadataPatch(checkRunID *int64, logURL string) map[string]any {
	return map[string]any{
		"metadata": map[string]any{
			"labels": map[string]string{
				keys.CheckRunID: strconv.FormatInt(*checkRunID, 10),
			},
			"annotations": map[string]string{
				keys.LogURL:     logURL,
				keys.CheckRunID: strconv.FormatInt(*checkRunID, 10),
			},
		},
	}
}

// createStatusCommit use the classic/old statuses API which is available when we
// don't have a github app token.
func (v *Provider) createStatusCommit(ctx context.Context, runevent *info.Event, status provider.StatusOpts) error {
	var err error
	now := time.Now()
	switch status.Conclusion {
	case "neutral":
		status.Conclusion = "success" // We don't have a choice than setting as success, no pending here.
	case "pending":
		if status.Title != "" {
			status.Conclusion = "pending"
		}
	}
	if status.Status == "in_progress" {
		status.Conclusion = "pending"
	}

	ghstatus := &github.RepoStatus{
		State:       github.Ptr(status.Conclusion),
		TargetURL:   github.Ptr(status.DetailsURL),
		Description: github.Ptr(status.Title),
		Context:     github.Ptr(provider.GetCheckName(status, v.pacInfo)),
		CreatedAt:   &github.Timestamp{Time: now},
	}

	if _, _, err := v.Client().Repositories.CreateStatus(ctx,
		runevent.Organization, runevent.Repository, runevent.SHA, ghstatus); err != nil {
		return err
	}
	eventType := triggertype.IsPullRequestType(runevent.EventType)
	if opscomments.IsAnyOpsEventType(eventType.String()) {
		eventType = triggertype.PullRequest
	}

	var commentStrategy string
	if v.repo != nil && v.repo.Spec.Settings != nil && v.repo.Spec.Settings.Github != nil {
		commentStrategy = v.repo.Spec.Settings.Github.CommentStrategy
	}

	switch commentStrategy {
	case "disable_all":
		v.Logger.Warn("github: comments related to PipelineRuns status have been disabled for Github pull requests")
		return nil
	default:
		if (status.Status == "completed" || (status.Status == "queued" && status.Title == pendingApproval)) &&
			status.Text != "" && eventType == triggertype.PullRequest {
			_, _, err = v.Client().Issues.CreateComment(ctx, runevent.Organization, runevent.Repository,
				runevent.PullRequestNumber,
				&github.IssueComment{
					Body: github.Ptr(fmt.Sprintf("%s<br>%s", status.Summary, status.Text)),
				},
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *Provider) CreateStatus(ctx context.Context, runevent *info.Event, statusOpts provider.StatusOpts) error {
	if v.ghClient == nil {
		return fmt.Errorf("cannot set status on github no token or url set")
	}

	// If the request comes from a bot user, skip setting the status and just log the event silently
	if statusOpts.AccessDenied && v.userType == botType {
		return nil
	}

	switch statusOpts.Conclusion {
	case "success":
		statusOpts.Title = "Success"
		statusOpts.Summary = "has <b>successfully</b> validated your commit."
	case "failure":
		statusOpts.Title = "Failed"
		statusOpts.Summary = "has <b>failed</b>."
	case "pending":
		// for concurrency set title as pending
		if statusOpts.Title == "" {
			statusOpts.Title = "Pending"
			statusOpts.Summary = "is skipping this commit."
		} else {
			// for unauthorized user set title as Pending approval
			statusOpts.Summary = "is waiting for approval."
		}
	case "cancelled":
		statusOpts.Title = "Cancelled"
		statusOpts.Summary = "has been <b>cancelled</b>."
	case "neutral":
		if statusOpts.Title == "" {
			statusOpts.Title = "Unknown"
		}
		statusOpts.Summary = "<b>Completed</b>"
	}

	if statusOpts.Status == "in_progress" {
		statusOpts.Title = "CI has Started"
		statusOpts.Summary = "is running."
	}

	onPr := ""
	if statusOpts.OriginalPipelineRunName != "" {
		onPr = "/" + statusOpts.OriginalPipelineRunName
	}
	statusOpts.Summary = fmt.Sprintf("%s%s %s", v.pacInfo.ApplicationName, onPr, statusOpts.Summary)
	// If we have an installationID which mean we have a github apps and we can use the checkRun API
	if runevent.InstallationID > 0 {
		return v.getOrUpdateCheckRunStatus(ctx, runevent, statusOpts)
	}

	// Otherwise use the update status commit API
	return v.createStatusCommit(ctx, runevent, statusOpts)
}
