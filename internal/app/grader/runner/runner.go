package runner

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"grader/pkg/logger"
	"grader/pkg/workerpool"
	"io"
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
	Path string `json:"path" validate:"required,file"`
}

func CheckAssessmentJob(assessment Assessment) workerpool.Job {
	return func(ctx context.Context) error {
		l := logger.Global().WithComponent("CheckAssessmentJob")

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
		//io.Copy(os.Stdout, out)

		l.Debug().Str("container_image", assessment.ContainerImage).Msg("Creating container")
		resp, err := cli.ContainerCreate(ctx, &container.Config{
			Image:        assessment.ContainerImage,
			Cmd:          []string{assessment.PartID},
			Tty:          true,
			AttachStdout: true,
			AttachStderr: true,
		}, &container.HostConfig{
			Binds: []string{},
		}, nil, nil, "grader_"+assessment.TaskID.String())
		if err != nil {
			return fmt.Errorf("container create: %w", err)
		}

		if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			return fmt.Errorf("container start: %w", err)
		}

		statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
		select {
		case err := <-errCh:
			if err != nil {
				return fmt.Errorf("container finish: %w", err)
			}
		case <-statusCh:
		}

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

		buf, err := io.ReadAll(out)
		if err != nil {
			return fmt.Errorf("read out: %w", err)
		}
		l.Debug().Bytes("out", buf)

		return nil
	}
}
