package runner

import (
	"bufio"
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
)

type Assessment struct {
	TaskID         uuid.UUID        `json:"-"`
	ContainerImage string           `json:"container_image" validate:"required"`
	PartID         string           `json:"part_id" validate:"required"`
	PostbackURL    string           `json:"postback_url" validate:"required,url"`
	Files          []AssessmentFile `json:"files" validate:"required"`
}

type AssessmentFile struct {
	Name string `json:"name" validate:"required"`
	URL  string `json:"url" validate:"required,url"`
}

func CheckAssessmentJob(assessment Assessment) workerpool.Job {
	return func(ctx context.Context) error {
		l := logger.Global().WithComponent("CheckAssessmentJob")

		l.Debug().Msg("Creating temporary dir")
		tempDir, err := ioutil.TempDir("", "assessment*")
		if err != nil {
			return fmt.Errorf("temp tempDir: %w", err)
		}
		defer func(path string) {
			_ = os.RemoveAll(path)
		}(tempDir)
		l.Debug().Str("path", tempDir).Msg("Using dir")

		if err := fetchAssessmentFiles(ctx, tempDir, assessment.Files); err != nil {
			return fmt.Errorf("fetch files: %w", err)
		}

		l.Debug().Msg("Creating client")
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return fmt.Errorf("new client: %w", err)
		}

		l.Debug().Str("container_image", assessment.ContainerImage).Msg("Pulling image")
		out, err := cli.ImagePull(ctx, assessment.ContainerImage, types.ImagePullOptions{})
		if err != nil {
			return fmt.Errorf("image pull: %w", err)
		}
		defer func(out io.ReadCloser) {
			_ = out.Close()
		}(out)
		io.Copy(os.Stdout, out)

		l.Debug().Str("container_image", assessment.ContainerImage).Msg("Creating container")
		resp, err := cli.ContainerCreate(ctx, &container.Config{
			Image:        assessment.ContainerImage,
			Cmd:          []string{"test", fmt.Sprintf("PART_ID=%s", assessment.PartID)},
			Tty:          true,
			AttachStdout: true,
			AttachStderr: true,
		}, &container.HostConfig{
			Binds: []string{
				fmt.Sprintf("%s:/app/submission", tempDir),
			},
		}, nil, nil, "grader_"+assessment.TaskID.String())
		if err != nil {
			return fmt.Errorf("container create: %w", err)
		}

		l.Debug().
			Str("container_image", assessment.ContainerImage).
			Str("container_id", resp.ID).
			Msg("Starting container")
		if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			return fmt.Errorf("container start: %w", err)
		}

		l.Debug().
			Str("container_image", assessment.ContainerImage).
			Str("container_id", resp.ID).
			Msg("Waiting for container to finish")
		statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
		select {
		case err := <-errCh:
			if err != nil {
				return fmt.Errorf("container finish: %w", err)
			}
		case s := <-statusCh:
			l.Debug().
				Str("container_image", assessment.ContainerImage).
				Str("container_id", resp.ID).
				Msgf("status %+v", s)
		}

		l.Debug().
			Str("container_image", assessment.ContainerImage).
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

		l.Debug().
			Str("container_image", assessment.ContainerImage).
			Str("container_id", resp.ID).
			Msg(buf.String())

		scanner := bufio.NewScanner(out)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}

		return nil
	}
}

// fetchAssessmentFiles to target directory
func fetchAssessmentFiles(ctx context.Context, targetDir string, files []AssessmentFile) error {
	if len(files) == 0 {
		return nil
	}

	g, ctx := errgroup.WithContext(ctx)
	for _, f := range files {
		g.Go(func() error {
			return fetchFile(f.URL, targetDir+"/"+f.Name)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("errgroup: %w", err)
	}

	return nil
}

// fetchFile from URL and save to path
func fetchFile(URL string, path string) error {
	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("file create: %w", err)
	}
	defer func(out *os.File) {
		_ = out.Close()
	}(out)

	resp, err := http.Get(URL)
	if err != nil {
		return fmt.Errorf("http get: %w", err)
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
