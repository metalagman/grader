package runner

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
	"grader/pkg/logger"
	"grader/pkg/workerpool"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type Submission struct {
	TaskID         uuid.UUID        `json:"-"`
	ContainerImage string           `json:"container_image" validate:"required"`
	PartID         string           `json:"part_id" validate:"required"`
	PostbackURL    string           `json:"postback_url" validate:"required,url"`
	Files          []SubmissionFile `json:"files" validate:"required"`
}

type SubmissionFile struct {
	Name string `json:"name" validate:"required"`
	URL  string `json:"url" validate:"required,url"`
}

type ContainerError struct {
	Output     string
	StatusCode int64
}

func (c ContainerError) Error() string {
	return c.Output
}

func CheckSubmissionJob(submission Submission) workerpool.Job {
	return func(ctx context.Context) error {
		l := logger.Global().WithComponent("CheckSubmissionJob")

		l.Debug().Msg("Creating temporary dir")
		tempDir, err := ioutil.TempDir("", "submission*")
		if err != nil {
			return fmt.Errorf("temp dir: %w", err)
		}
		defer func(path string) {
			_ = os.RemoveAll(path)
		}(tempDir)
		l.Debug().Str("path", tempDir).Msg("Using temporary dir")

		if err := fetchSubmissionFiles(ctx, tempDir, submission.Files); err != nil {
			return fmt.Errorf("fetch files: %w", err)
		}

		r := SubmissionResult{
			TaskID: submission.TaskID,
		}

		if err := runContainer(ctx, l, submission, tempDir); err != nil {
			r.Text = err.Error()
		} else {
			r.Text = "OK"
			r.Pass = true
		}

		sendCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := sendResult(sendCtx, submission.PostbackURL, r); err != nil {
			return fmt.Errorf("send result: %w", err)
		}

		return nil
	}
}

func runContainer(ctx context.Context, l logger.Logger, submission Submission, submissionDir string) error {
	l.Debug().Msg("Creating client")
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("new client: %w", err)
	}

	l.Debug().Str("container_image", submission.ContainerImage).Msg("Pulling image")
	out, err := cli.ImagePull(ctx, submission.ContainerImage, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("image pull: %w", err)
	}
	defer func(out io.ReadCloser) {
		_ = out.Close()
	}(out)
	//io.Copy(os.Stdout, out)

	l.Debug().Str("container_image", submission.ContainerImage).Msg("Creating container")
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        submission.ContainerImage,
		Cmd:          []string{"test", fmt.Sprintf("PART_ID=%s", submission.PartID)},
		Tty:          true,
		AttachStdout: true,
		AttachStderr: true,
	}, &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/app/submission", submissionDir),
		},
	}, nil, nil, "grader_"+submission.TaskID.String())
	if err != nil {
		return fmt.Errorf("container create: %w", err)
	}

	l.Debug().
		Str("container_image", submission.ContainerImage).
		Str("container_id", resp.ID).
		Msg("Starting container")
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("container start: %w", err)
	}

	l.Debug().
		Str("container_image", submission.ContainerImage).
		Str("container_id", resp.ID).
		Msg("Waiting for container to finish")
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)

	var s container.ContainerWaitOKBody

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("container finish: %w", err)
		}
	case s = <-statusCh:
		l.Debug().
			Str("container_image", submission.ContainerImage).
			Str("container_id", resp.ID).
			Int64("container_status", s.StatusCode).
			Msg("Container failed")
	}

	l.Debug().
		Str("container_image", submission.ContainerImage).
		Str("container_id", resp.ID).
		Msg("Reading logs")
	out, err = cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: false,
	})
	if err != nil {
		return fmt.Errorf("container out: %w", err)
	}
	defer func(logs io.ReadCloser) {
		_ = logs.Close()
	}(out)

	buf := new(strings.Builder)
	_, err = io.Copy(buf, out)
	if err != nil {
		return fmt.Errorf("logs: %w", err)
	}
	result := buf.String()

	if s.StatusCode > 0 {
		return ContainerError{result, s.StatusCode}
	}

	return nil

}

// fetchSubmissionFiles to target directory
func fetchSubmissionFiles(ctx context.Context, targetDir string, files []SubmissionFile) error {
	if len(files) == 0 {
		return nil
	}

	g, ctx := errgroup.WithContext(ctx)
	for _, f := range files {
		f := f
		g.Go(func() error {
			return fetchFile(ctx, f.URL, targetDir+"/"+f.Name)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("errgroup: %w", err)
	}

	return nil
}

// fetchFile from URL and save to path
func fetchFile(ctx context.Context, URL string, path string) error {
	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("file create: %w", err)
	}
	defer func(out *os.File) {
		_ = out.Close()
	}(out)

	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	req = req.WithContext(ctx)

	cl := http.DefaultClient
	resp, err := cl.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("copy: %w", err)
	}

	return nil
}
