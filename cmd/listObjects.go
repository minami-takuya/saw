package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag/v2"

	"github.com/minami-takuya/saw/internal"
)

type ListObjectsWriter interface {
	WriteObject(out io.Writer, bucket string, object types.Object) error
}

type ListObjectsPlainWriter struct{}

func (w *ListObjectsPlainWriter) WriteObject(out io.Writer, bucket string, object types.Object) error {
	_, err := fmt.Fprintf(out, "%s/%s\n", bucket, *object.Key)
	return err
}

type ListObjectsCwlS3EventWriter struct {
	DetailType string
	Reason     string
}

func (w *ListObjectsCwlS3EventWriter) WriteObject(out io.Writer, bucket string, object types.Object) error {
	detail := internal.S3EventDetail{
		Version: "0",
		Bucket: internal.S3Bucket{
			Name: bucket,
		},
		Object: internal.S3Object{
			Key:  *object.Key,
			Size: *object.Size,
			ETag: *object.ETag,
		},
		RequestID: "",
		Requester: "",
		SourceIP:  "",
		Reason:    w.Reason,
	}
	b, err := json.Marshal(detail)
	if err != nil {
		return err
	}

	evt := events.CloudWatchEvent{
		Version:    "0",
		ID:         uuid.New().String(),
		DetailType: w.DetailType,
		Source:     "aws.s3",
		Time:       time.Now(),
		AccountID:  "",
		Region:     "",
		Resources: []string{
			fmt.Sprintf("arn:aws:s3:::%s", bucket),
		},
		Detail: b,
	}
	t, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "%s\n", string(t))
	if err != nil {
		return err
	}
	return nil
}

//go:generate go run github.com/dmarkham/enumer -type=ListObjectsFormat -json -text -yaml -trimprefix=ListObjectsFormat -transform=kebab
type ListObjectsFormat enumflag.Flag

const (
	ListObjectsFormatPlain ListObjectsFormat = iota
	ListObjectsFormatCwlS3Event
)

var ListObjectsFormatStringMap = map[ListObjectsFormat][]string{
	ListObjectsFormatPlain:      {"plain"},
	ListObjectsFormatCwlS3Event: {"cwl-s3-event"},
}

type ListObjectsOpts struct {
	Bucket string

	Prefixes  []string
	Delimiter string

	Format ListObjectsFormat

	// for cwl-s3-event
	DetailType string
	Reason     string
}

var (
	listObjectsOpts = &ListObjectsOpts{
		Bucket:     "",
		Prefixes:   []string{""},
		Delimiter:  "",
		Format:     ListObjectsFormatPlain,
		DetailType: "Object Created",
		Reason:     "PutObject",
	}
)

// listObjectsCmd represents the listObjects command
var listObjectsCmd = &cobra.Command{
	Use:   "list-objects",
	Short: "list objects in an s3 bucket",
	RunE: func(cmd *cobra.Command, args []string) error {
		if listObjectsOpts.Bucket == "" {
			return cmd.Help()
		}

		cfg, err := config.LoadDefaultConfig(cmd.Context())
		if err != nil {
			return err
		}
		client := s3.NewFromConfig(cfg)
		wg := new(sync.WaitGroup)

		writer := ListObjectsWriter(&ListObjectsPlainWriter{})
		if listObjectsOpts.Format == ListObjectsFormatCwlS3Event {
			writer = ListObjectsWriter(&ListObjectsCwlS3EventWriter{
				DetailType: listObjectsOpts.DetailType,
				Reason:     listObjectsOpts.Reason,
			})
		}

		prefixes := internal.SliceCh(listObjectsOpts.Prefixes)
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				for prefix := range prefixes {
					paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
						Bucket:    &listObjectsOpts.Bucket,
						Prefix:    &prefix,
						Delimiter: &listObjectsOpts.Delimiter,
					})

					for paginator.HasMorePages() {
						output, err := paginator.NextPage(cmd.Context())
						if err != nil {
							slog.Error("failed to list objects", slog.Any("error", err))
							return
						}

						for _, object := range output.Contents {
							if err := writer.WriteObject(cmd.OutOrStdout(), listObjectsOpts.Bucket, object); err != nil {
								slog.Error("failed to write object", slog.Any("error", err))
								return
							}
						}
					}
				}
			}()
		}
		wg.Wait()
		return nil
	},
}

func init() {
	s3Cmd.AddCommand(listObjectsCmd)

	listObjectsCmd.Flags().StringVarP(&listObjectsOpts.Bucket, "bucket", "b", listObjectsOpts.Bucket, "the bucket to list objects from")
	listObjectsCmd.Flags().StringSliceVarP(&listObjectsOpts.Prefixes, "prefix", "p", listObjectsOpts.Prefixes, "the prefix to filter objects by")
	listObjectsCmd.Flags().StringVarP(&listObjectsOpts.Delimiter, "delimiter", "d", listObjectsOpts.Delimiter, "the delimiter to use when listing objects")
	listObjectsCmd.Flags().VarP(enumflag.New(&listObjectsOpts.Format, "format", ListObjectsFormatStringMap, enumflag.EnumCaseInsensitive), "format", "f", "the output format")
	listObjectsCmd.Flags().StringVar(&listObjectsOpts.DetailType, "detail-type", listObjectsOpts.DetailType, "the detail type for cwl-s3-event format")
	listObjectsCmd.Flags().StringVar(&listObjectsOpts.Reason, "reason", listObjectsOpts.Reason, "the reason for cwl-s3-event format")
}
